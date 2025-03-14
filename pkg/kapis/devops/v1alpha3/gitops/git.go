package gitops

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/emirpasic/gods/trees/binaryheap"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	commitgraphfmt "github.com/go-git/go-git/v5/plumbing/format/commitgraph/v2"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/object/commitgraph"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/kubesphere/ks-devops/pkg/kapis/devops/v1alpha3/utils"
	"k8s.io/klog/v2"
)

type gitRepoService struct {
	author *object.Signature
	repo   *git.Repository
	auth   transport.AuthMethod
	// InsecureSkipTLS skips ssl verify if protocol is https
	insecureSkipTLS bool
	// CABundle specify additional ca bundle with system cert pool
	caBundle []byte
	// ProxyOptions provides info required for connecting to a proxy.
	proxyOptions transport.ProxyOptions
	newFilePerm  os.FileMode
}

func (s *gitRepoService) UploadFiles(ctx context.Context, input *UploadFilesInput) (*UploadFilesOutput, error) {
	if len(input.Files) == 0 {
		return nil, os.ErrInvalid
	}

	wt, err := s.repo.Worktree()
	if err != nil {
		return nil, err
	}

	root := wt.Filesystem.Root()
	uploadRoot := s.getUploadRoot(root)

	deleteFilesFunc := func(files []string) {
		for _, file := range files {
			err := os.Remove(file)
			if err != nil {
				klog.ErrorS(err, "delete file failed", "file", file)
			}
		}
	}

	uploadFilesFunc := func() ([]string, error) {
		var filePaths []string
		for _, file := range input.Files {
			filePath := filepath.Join(uploadRoot, file.Name)
			filePaths = append(filePaths, filePath)
			err = os.RemoveAll(filePath)
			if err != nil {
				return filePaths, err
			}
			err = os.MkdirAll(filepath.Dir(filePath), s.newFilePerm)
			if err != nil {
				return filePaths, err
			}
			err = os.WriteFile(filePath, file.Data, s.newFilePerm)
			if err != nil {
				return filePaths, err
			}
		}
		return filePaths, nil
	}

	filePaths, err := uploadFilesFunc()
	if err != nil {
		deleteFilesFunc(filePaths)
		return nil, err
	}

	// delete files after 1 hour
	timer := time.NewTimer(time.Hour)
	go func(filePaths []string) {
		<-timer.C
		deleteFilesFunc(filePaths)
	}(filePaths)

	out := &UploadFilesOutput{}
	return out, nil
}

func (s *gitRepoService) GetBranch(ctx context.Context, input *GetBranchInput) (*GetBranchOutput, error) {
	branchesOut, err := s.ListBranches(ctx, &ListBranchesInput{
		Options: &ListOptions{
			Page:  1,
			Limit: 1000,
		},
		Remote:   input.Remote,
		WithHead: true,
	})
	if err != nil {
		return nil, err
	}

	for _, branch := range branchesOut.Items {
		if branch.Name == input.Branch {
			return &GetBranchOutput{
				Branch: branch,
			}, nil
		}
	}

	return nil, ErrBranchNotFound
}

func (s *gitRepoService) GetCommit(ctx context.Context, input *GetCommitInput) (*GetCommitOutput, error) {
	commit, err := s.repo.CommitObject(plumbing.NewHash(input.Commit))
	if err != nil {
		return nil, err
	}
	out := &GetCommitOutput{
		Commit: convertCommit(commit),
	}
	return out, nil
}

func (s *gitRepoService) GetConfig(ctx context.Context, input *GetConfigInput) (*GetConfigOutput, error) {
	cfg, err := s.repo.Config()
	if err != nil {
		return nil, err
	}
	out := &GetConfigOutput{
		Config: cfg,
	}
	return out, nil
}

func (s *gitRepoService) UpdateConfig(ctx context.Context, input *UpdateConfigInput) (*UpdateConfigOutput, error) {
	if input.Config == nil {
		return nil, os.ErrInvalid
	}
	err := s.repo.SetConfig(input.Config)
	if err != nil {
		return nil, err
	}
	getConfigOut, err := s.GetConfig(ctx, &GetConfigInput{})
	if err != nil {
		return nil, err
	}
	out := &UpdateConfigOutput{
		Config: getConfigOut.Config,
	}
	return out, nil
}

func (s *gitRepoService) DeleteClone(ctx context.Context, input *DeleteCloneInput) (*DeleteCloneOutput, error) {
	workTree, err := s.repo.Worktree()
	if err != nil {
		return nil, err
	}
	root := workTree.Filesystem.Root()
	err = os.RemoveAll(root)
	if err != nil {
		return nil, err
	}
	out := &DeleteCloneOutput{}
	return out, nil
}

func (s *gitRepoService) ListFiles(ctx context.Context, input *ListFilesInput) (*ListFilesOutput, error) {
	if len(input.Dir) == 0 || len(input.Branch) == 0 {
		return nil, os.ErrInvalid
	}

	treePath := input.Dir
	// currently we only support get last commits for files under specified directory(treePath)
	if !strings.HasSuffix(treePath, "/") {
		return nil, os.ErrInvalid
	}

	treePath = strings.TrimSuffix(strings.TrimPrefix(treePath, "/"), "/")

	coOut, err := s.CheckOutBranch(ctx, &CheckOutBranchInput{
		Branch: input.Branch,
		Force:  true,
	})
	if err != nil {
		return nil, err
	}

	wt := coOut.WorkTree

	head, err := s.repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := s.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	if treePath != "" {
		tree, err = tree.Tree(treePath)
		if err != nil {
			return nil, err
		}
	}

	fileCommits := []*FileCommitInfo{
		// put the treePath as first element
		{
			FileInfo: FileInfo{
				FileNameData: FileNameData{
					Name: treePath,
				},
				IsDir: true,
			},
		},
	}

	var paths []string
	for _, entry := range tree.Entries {
		paths = append(paths, entry.Name)

		if entry.Mode == filemode.Regular || entry.Mode == filemode.Executable || entry.Mode == filemode.Symlink {
			fileOut, err := s.GetFile(ctx, &GetFileInput{
				Commit:          commit.Hash.String(),
				File:            path.Join(treePath, entry.Name),
				WithFileContent: input.WithFileContent,
			})
			if err != nil {
				return nil, err
			}
			info := &FileCommitInfo{
				FileInfo: *fileOut.File,
			}
			fileCommits = append(fileCommits, info)
		} else if entry.Mode == filemode.Dir {
			info := &FileCommitInfo{
				FileInfo: FileInfo{
					FileNameData: FileNameData{
						Name: entry.Name,
					},
					IsDir: true,
				},
			}
			fileCommits = append(fileCommits, info)
		}
	}

	if input.WithLastCommit {
		fs := wt.Filesystem

		commitNodeIndex, file := getCommitNodeIndex(s.repo, fs)
		if file != nil {
			defer file.Close()
		}

		commitNode, err := commitNodeIndex.Get(head.Hash())
		if err != nil {
			return nil, err
		}

		revs, err := getLastCommitForPaths(commitNode, treePath, paths)
		if err != nil {
			return nil, err
		}

		var lastCommitTime time.Time
		for k, v := range fileCommits {
			if k == 0 {
				continue
			}
			c, ok := revs[v.Name]
			if ok {
				v.Commit = convertCommit(c)

				// update last commit for treePath
				if v.Commit.Committer.When.After(lastCommitTime) {
					lastCommitTime = v.Commit.Committer.When
					fileCommits[0].Commit = v.Commit
				}
			}
		}
	}

	out := &ListFilesOutput{
		Items:      fileCommits,
		TotalItems: len(fileCommits),
		Options: &ListOptions{
			Page:  1,
			Limit: 10000,
		},
	}

	return out, nil
}

func getCommitNodeIndex(r *git.Repository, fs billy.Filesystem) (commitgraph.CommitNodeIndex, io.ReadCloser) {
	file, err := fs.Open(filepath.Join("objects", "info", "commit-graph"))
	if err == nil {
		index, err := commitgraphfmt.OpenFileIndex(file)
		if err == nil {
			return commitgraph.NewGraphCommitNodeIndex(index, r.Storer), file
		}
		_ = file.Close()
	}

	return commitgraph.NewObjectCommitNodeIndex(r.Storer), nil
}

type commitAndPaths struct {
	commit commitgraph.CommitNode
	// Paths that are still on the branch represented by commit
	paths []string
	// Set of hashes for the paths
	hashes map[string]plumbing.Hash
}

func getCommitTree(c commitgraph.CommitNode, treePath string) (*object.Tree, error) {
	tree, err := c.Tree()
	if err != nil {
		return nil, err
	}

	// Optimize deep traversals by focusing only on the specific tree
	if treePath != "" {
		tree, err = tree.Tree(treePath)
		if err != nil {
			return nil, err
		}
	}

	return tree, nil
}

func getFileHashes(c commitgraph.CommitNode, treePath string, paths []string) (map[string]plumbing.Hash, error) {
	tree, err := getCommitTree(c, treePath)
	if errors.Is(err, object.ErrDirectoryNotFound) {
		// The whole tree didn't exist, so return empty map
		return make(map[string]plumbing.Hash), nil
	}
	if err != nil {
		return nil, err
	}

	hashes := make(map[string]plumbing.Hash)
	for _, path := range paths {
		if path != "" {
			entry, err := tree.FindEntry(path)
			if err == nil {
				hashes[path] = entry.Hash
			}
		} else {
			hashes[path] = tree.Hash
		}
	}

	return hashes, nil
}

func getLastCommitForPaths(c commitgraph.CommitNode, treePath string, paths []string) (map[string]*object.Commit, error) {
	// We do a tree traversal with nodes sorted by commit time
	heap := binaryheap.NewWith(func(a, b interface{}) int {
		if a.(*commitAndPaths).commit.CommitTime().Before(b.(*commitAndPaths).commit.CommitTime()) {
			return 1
		}
		return -1
	})

	resultNodes := make(map[string]commitgraph.CommitNode)
	initialHashes, err := getFileHashes(c, treePath, paths)
	if err != nil {
		return nil, err
	}

	// Start search from the root commit and with full set of paths
	heap.Push(&commitAndPaths{c, paths, initialHashes})

	for {
		cIn, ok := heap.Pop()
		if !ok {
			break
		}
		current := cIn.(*commitAndPaths)

		// Load the parent commits for the one we are currently examining
		numParents := current.commit.NumParents()
		var parents []commitgraph.CommitNode
		for i := 0; i < numParents; i++ {
			parent, err := current.commit.ParentNode(i)
			if err != nil {
				break
			}
			parents = append(parents, parent)
		}

		// Examine the current commit and set of interesting paths
		pathUnchanged := make([]bool, len(current.paths))
		parentHashes := make([]map[string]plumbing.Hash, len(parents))
		for j, parent := range parents {
			parentHashes[j], err = getFileHashes(parent, treePath, current.paths)
			if err != nil {
				break
			}

			for i, path := range current.paths {
				if parentHashes[j][path] == current.hashes[path] {
					pathUnchanged[i] = true
				}
			}
		}

		var remainingPaths []string
		for i, path := range current.paths {
			// The results could already contain some newer change for the same path,
			// so don't override that and bail out on the file early.
			if resultNodes[path] == nil {
				if pathUnchanged[i] {
					// The path existed with the same hash in at least one parent so it could
					// not have been changed in this commit directly.
					remainingPaths = append(remainingPaths, path)
				} else {
					// There are few possible cases how can we get here:
					// - The path didn't exist in any parent, so it must have been created by
					//   this commit.
					// - The path did exist in the parent commit, but the hash of the file has
					//   changed.
					// - We are looking at a merge commit and the hash of the file doesn't
					//   match any of the hashes being merged. This is more common for directories,
					//   but it can also happen if a file is changed through conflict resolution.
					resultNodes[path] = current.commit
				}
			}
		}

		if len(remainingPaths) > 0 {
			// Add the parent nodes along with remaining paths to the heap for further
			// processing.
			for j, parent := range parents {
				// Combine remainingPath with paths available on the parent branch
				// and make union of them
				remainingPathsForParent := make([]string, 0, len(remainingPaths))
				newRemainingPaths := make([]string, 0, len(remainingPaths))
				for _, path := range remainingPaths {
					if parentHashes[j][path] == current.hashes[path] {
						remainingPathsForParent = append(remainingPathsForParent, path)
					} else {
						newRemainingPaths = append(newRemainingPaths, path)
					}
				}

				if remainingPathsForParent != nil {
					heap.Push(&commitAndPaths{parent, remainingPathsForParent, parentHashes[j]})
				}

				if len(newRemainingPaths) == 0 {
					break
				} else {
					remainingPaths = newRemainingPaths
				}
			}
		}
	}

	// Post-processing
	result := make(map[string]*object.Commit)
	for path, commitNode := range resultNodes {
		var err error
		result[path], err = commitNode.Commit()
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *gitRepoService) CleanAndPull(ctx context.Context, input *CleanAndPullInput) (*CleanAndPullOutput, error) {
	w := input.WorkTree
	if w == nil {
		return nil, os.ErrInvalid
	}
	head, err := s.repo.Head()
	if err != nil {
		return nil, err
	}

	klog.InfoS("run 'git clean -f -d . && git reset --hard HEAD && git pull origin'")
	err = w.Clean(&git.CleanOptions{Dir: true})
	if err != nil {
		return nil, err
	}
	err = w.Reset(&git.ResetOptions{
		Commit: head.Hash(),
		Mode:   git.HardReset,
	})
	if err != nil {
		return nil, err
	}

	branchRefName := plumbing.NewBranchReferenceName(input.Branch)
	err = w.Pull(&git.PullOptions{
		RemoteName:      "origin",
		ReferenceName:   branchRefName,
		SingleBranch:    true,
		Force:           true,
		Auth:            s.auth,
		Progress:        os.Stdout,
		InsecureSkipTLS: s.insecureSkipTLS,
		CABundle:        s.caBundle,
		ProxyOptions:    s.proxyOptions,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil, err
	}

	out := &CleanAndPullOutput{}

	return out, nil
}

func (s *gitRepoService) CommitAndPush(ctx context.Context, input *CommitAndPushInput) (*CommitAndPushOutput, error) {
	if len(input.Message) == 0 || input.WorkTree == nil {
		return nil, os.ErrInvalid
	}
	w := input.WorkTree

	_, err := w.Add(".")
	if err != nil {
		return nil, err
	}

	status, err := w.Status()
	if err != nil {
		return nil, err
	}

	if len(status) == 0 {
		return nil, ErrWorkTreeClean
	}

	author := s.author
	if author == nil {
		getConfigOut, err := s.GetConfig(ctx, &GetConfigInput{})
		if err != nil {
			return nil, err
		}
		author = &object.Signature{
			Name:  getConfigOut.Config.Author.Name,
			Email: getConfigOut.Config.Author.Email,
		}
	}
	author.When = time.Now()

	commitMessage := input.Message
	if input.SignOff {
		// Create a sign-off line and append it to the commit message.
		signOff := fmt.Sprintf("Signed-off-by: %s", author.Name)
		if author.Email != "" {
			signOff = fmt.Sprintf("%s <%s>", signOff, author.Email)
		}
		commitMessage = fmt.Sprintf("%s\n\n%s", commitMessage, signOff)
	}

	// Commit the changes with the message and author information.
	hash, err := w.Commit(commitMessage, &git.CommitOptions{
		All:    true,
		Author: author,
	})
	if err != nil {
		return nil, err
	}

	err = s.repo.Push(&git.PushOptions{
		RemoteName:      "origin",
		Auth:            s.auth,
		Progress:        os.Stdout,
		Force:           true,
		InsecureSkipTLS: s.insecureSkipTLS,
		CABundle:        s.caBundle,
		Atomic:          true,
		ProxyOptions:    s.proxyOptions,
	})
	if err != nil {
		return nil, err
	}

	commit, err := s.repo.CommitObject(hash)
	if err != nil {
		return nil, err
	}
	out := &CommitAndPushOutput{
		Commit: convertCommit(commit),
	}

	return out, nil
}

func (s *gitRepoService) fetchOrigin(refSpecStr string) error {
	remote, err := s.repo.Remote("origin")
	if err != nil {
		return err
	}

	var refSpecs []config.RefSpec
	if refSpecStr != "" {
		refSpecs = []config.RefSpec{config.RefSpec(refSpecStr)}
	}

	err = remote.Fetch(&git.FetchOptions{
		RefSpecs:        refSpecs,
		Auth:            s.auth,
		Progress:        os.Stdout,
		Force:           true,
		InsecureSkipTLS: s.insecureSkipTLS,
		CABundle:        s.caBundle,
		ProxyOptions:    s.proxyOptions,
	})
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			klog.InfoS("refs already up to date")
		} else {
			return fmt.Errorf("fetch origin failed: %v", err)
		}
	}

	return nil
}

func (s *gitRepoService) CheckOutBranch(ctx context.Context, input *CheckOutBranchInput) (*CheckOutBranchOutput, error) {
	if len(input.Branch) == 0 {
		return nil, os.ErrInvalid
	}
	branchRefName := plumbing.NewBranchReferenceName(input.Branch)
	branchCoOpts := git.CheckoutOptions{
		Branch: branchRefName,
		Force:  input.Force,
	}
	w, err := s.repo.Worktree()
	if err != nil {
		return nil, err
	}
	err = w.Checkout(&branchCoOpts)
	if err != nil {
		klog.Warningf("local checkout of branch '%s' failed, will attempt to fetch remote branch of same name.", input.Branch)
		klog.Warningf("like `git checkout <branch>` defaulting to `git checkout -b <branch> --track <remote>/<branch>`")

		mirrorRemoteBranchRefSpec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", input.Branch, input.Branch)
		err = s.fetchOrigin(mirrorRemoteBranchRefSpec)
		if err != nil {
			return nil, err
		}

		err = w.Checkout(&branchCoOpts)
		if err != nil {
			return nil, err
		}
	}
	out := &CheckOutBranchOutput{
		WorkTree: w,
	}
	return out, nil
}

func (s *gitRepoService) ListBranches(ctx context.Context, input *ListBranchesInput) (*ListBranchesOutput, error) {
	out := &ListBranchesOutput{
		Options: input.Options,
	}

	var refs []*plumbing.Reference
	var err error

	if input.Remote {
		var remote *git.Remote
		remote, err = s.repo.Remote("origin")
		if err != nil {
			return nil, err
		}
		refs, err = remote.List(&git.ListOptions{
			Auth:            s.auth,
			InsecureSkipTLS: s.insecureSkipTLS,
			CABundle:        s.caBundle,
			ProxyOptions:    s.proxyOptions,
		})
		if err != nil {
			return nil, err
		}
	} else {
		var refIter storer.ReferenceIter
		refIter, err = s.repo.Branches()
		if err != nil {
			return nil, err
		}
		err = refIter.ForEach(func(ref *plumbing.Reference) error {
			refs = append(refs, ref)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	filteredRefs := utils.SortRefsByShortName(refs)
	if !input.WithHead {
		filteredRefs = make([]*plumbing.Reference, 0)
		for _, ref := range refs {
			if ref.Name().Short() != plumbing.HEAD.Short() {
				filteredRefs = append(filteredRefs, ref)
			}
		}
	}

	refs, out.TotalItems = utils.GetPage(filteredRefs, input.Options.Page, input.Options.Limit)
	for _, ref := range refs {
		hash := ref.Hash()
		if ref.Name().Short() == plumbing.HEAD.Short() {
			head, err := s.repo.Head()
			if err != nil {
				return nil, err
			}
			hash = head.Hash()
		}
		commit, err := s.repo.CommitObject(hash)
		if err != nil {
			return nil, err
		}
		branch := &BranchInfo{
			Ref:    ref.Name().String(),
			Name:   ref.Name().Short(),
			Commit: convertCommit(commit),
		}
		out.Items = append(out.Items, branch)
	}

	return out, err
}

func (s *gitRepoService) ListCommits(ctx context.Context, input *ListCommitsInput) (*ListCommitsOutput, error) {
	if len(input.Branch) == 0 {
		return nil, os.ErrInvalid
	}

	out := &ListCommitsOutput{
		Options: input.Options,
	}

	_, err := s.CheckOutBranch(ctx, &CheckOutBranchInput{
		Branch: input.Branch,
		Force:  true,
	})
	if err != nil {
		return nil, err
	}

	var fileName *string
	if input.FileName != "" {
		fileName = &input.FileName
	}
	commitIter, err := s.repo.Log(&git.LogOptions{
		FileName: fileName,
	})
	if err != nil {
		return nil, err
	}

	var commits []*object.Commit
	err = commitIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})
	if err != nil {
		return nil, err
	}
	commits, out.TotalItems = utils.GetPage(commits, input.Options.Page, input.Options.Limit)

	for _, c := range commits {
		ci := &CommitInfo{Commit: convertCommit(c)}
		out.Items = append(out.Items, ci)
	}

	return out, nil
}

func (s *gitRepoService) getUploadRoot(repoRoot string) string {
	return strings.TrimSuffix(repoRoot, "/") + "_upload_"
}

func (s *gitRepoService) AddFiles(ctx context.Context, input *AddFilesInput) (*AddFilesOutput, error) {
	if len(input.Branch) == 0 || len(input.Files) == 0 || len(input.Message) == 0 {
		return nil, os.ErrInvalid
	}
	coOut, err := s.CheckOutBranch(ctx, &CheckOutBranchInput{
		Branch: input.Branch,
		Force:  true,
	})
	if err != nil {
		return nil, err
	}
	w := coOut.WorkTree

	_, err = s.CleanAndPull(ctx, &CleanAndPullInput{
		WorkTree: w,
		Branch:   input.Branch,
	})
	if err != nil {
		return nil, err
	}

	root := w.Filesystem.Root()

	// if files have been uploaded, we need to read them first
	if input.Uploaded {
		uploadRoot := s.getUploadRoot(root)
		for _, file := range input.Files {
			filePath := filepath.Join(uploadRoot, file.Name)
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			file.Data = data
			_ = os.Remove(filePath)
		}
	}

	if input.Unpack {
		file := input.Files[0]
		filePath := w.Filesystem.Join(root, file.Name)
		err = utils.Unpack(file.Data, filePath, s.newFilePerm)
		if err != nil {
			return nil, err
		}
	} else {
		for _, file := range input.Files {
			_, err = w.Filesystem.Stat(file.Name)
			if err == nil && !input.Overwrite {
				continue
			}

			if len(file.OldName) > 0 {
				oldFilePath := w.Filesystem.Join(root, file.OldName)
				err = os.RemoveAll(oldFilePath)
				if err != nil {
					return nil, err
				}
			}

			filePath := w.Filesystem.Join(root, file.Name)
			err = os.RemoveAll(filePath)
			if err != nil {
				return nil, err
			}
			err = os.MkdirAll(filepath.Dir(filePath), s.newFilePerm)
			if err != nil {
				return nil, err
			}
			err = os.WriteFile(filePath, file.Data, s.newFilePerm)
			if err != nil {
				return nil, err
			}
		}
	}

	cpOut, err := s.CommitAndPush(ctx, &CommitAndPushInput{
		WorkTree: w,
		Message:  input.Message,
		SignOff:  true,
	})
	if err != nil {
		return nil, err
	}

	out := &AddFilesOutput{
		Commit: cpOut.Commit,
	}

	return out, nil
}

func (s *gitRepoService) DeleteFiles(ctx context.Context, input *DeleteFilesInput) (*DeleteFilesOutput, error) {
	if len(input.Branch) == 0 || len(input.Files) == 0 || len(input.Message) == 0 {
		return nil, os.ErrInvalid
	}
	coOut, err := s.CheckOutBranch(ctx, &CheckOutBranchInput{
		Branch: input.Branch,
		Force:  true,
	})
	if err != nil {
		return nil, err
	}
	w := coOut.WorkTree

	_, err = s.CleanAndPull(ctx, &CleanAndPullInput{
		WorkTree: w,
		Branch:   input.Branch,
	})
	if err != nil {
		return nil, err
	}

	for _, file := range input.Files {
		_, err = w.Remove(file)
		if err != nil {
			return nil, err
		}
	}

	cpOut, err := s.CommitAndPush(ctx, &CommitAndPushInput{
		WorkTree: w,
		Message:  input.Message,
		SignOff:  true,
	})
	if err != nil {
		return nil, err
	}

	out := &DeleteFilesOutput{
		Commit: cpOut.Commit,
	}

	return out, nil
}

func (s *gitRepoService) GetFile(ctx context.Context, input *GetFileInput) (*GetFileOutput, error) {
	if (len(input.Commit) == 0 && len(input.Branch) == 0) || len(input.File) == 0 {
		return nil, os.ErrInvalid
	}

	hash := plumbing.NewHash(input.Commit)
	if input.Commit == "" && input.Branch != "" {
		_, err := s.CheckOutBranch(ctx, &CheckOutBranchInput{
			Branch: input.Branch,
			Force:  false,
		})
		if err != nil {
			return nil, err
		}
		head, err := s.repo.Head()
		if err != nil {
			return nil, err
		}
		hash = head.Hash()
	}

	commit, err := s.repo.CommitObject(hash)
	if err != nil {
		return nil, err
	}

	baseName := path.Base(input.File)
	file, err := commit.File(strings.TrimPrefix(input.File, "/"))
	if err != nil {
		return nil, err
	}
	info := &FileInfo{
		FileNameData: FileNameData{
			Name: baseName,
		},
		Size: file.Size,
	}
	isBinary, err := file.IsBinary()
	if err != nil {
		return nil, err
	}
	info.IsBinary = isBinary
	info.IsSymlink = file.Mode == filemode.Symlink
	reader, err := file.Reader()
	if err != nil {
		return nil, err
	}
	if input.WithFileContent && file.Size <= UploadDownloadFileSizeLimit {
		str, err := file.Contents()
		if err != nil {
			return nil, err
		}
		info.Data = []byte(str)
	}

	out := &GetFileOutput{
		File:   info,
		Reader: reader,
	}

	return out, nil
}

var _ GitRepoService = &gitRepoService{}

func NewGitRepoService(opts *GitRepoOptions) GitRepoService {
	return &gitRepoService{
		author:          opts.author,
		repo:            opts.repo,
		auth:            opts.auth,
		insecureSkipTLS: opts.insecureSkipTLS,
		caBundle:        opts.caBundle,
		proxyOptions:    opts.proxyOptions,
		newFilePerm:     opts.newFilePerm,
	}
}
