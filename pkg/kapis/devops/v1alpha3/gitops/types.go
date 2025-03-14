package gitops

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/kubesphere/ks-devops/pkg/kapis/common"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/authentication/user"
)

var (
	ErrWorkTreeClean  = errors.New("nothing to commit, working tree clean")
	ErrBranchNotFound = errors.New("branch not found")
)

const (
	UploadDownloadFileSizeLimit = 1024 * 1024 * 10 // 10 MB
)

// Commit use most fields of object.Commit
type Commit struct {
	// Hash of the commit object.
	Hash string
	// Author is the original author of the commit.
	Author object.Signature
	// Committer is the one performing the commit, might be different from
	// Author.
	Committer object.Signature
	// MergeTag is the embedded tag object when a merge commit is created by
	// merging a signed tag.
	MergeTag string
	// Message is the commit message, contains arbitrary text.
	Message string
	// TreeHash is the hash of the root tree of the commit.
	TreeHash string
	// ParentHashes are the hashes of the parent commits of the commit.
	ParentHashes []string
}

func convertCommit(c *object.Commit) *Commit {
	if c == nil {
		return nil
	}
	var parentHashes []string
	for _, hash := range c.ParentHashes {
		parentHashes = append(parentHashes, hash.String())
	}
	return &Commit{
		Hash:         c.Hash.String(),
		Author:       c.Author,
		Committer:    c.Committer,
		MergeTag:     c.MergeTag,
		Message:      c.Message,
		TreeHash:     c.TreeHash.String(),
		ParentHashes: parentHashes,
	}
}

type ListResult[E any] struct {
	Items      []E          `json:"items"`
	TotalItems int          `json:"totalItems"`
	Options    *ListOptions `json:"options"`
}

type ListOptions struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func (o *ListOptions) Correct() *ListOptions {
	if o.Page <= 0 {
		o.Page = 1
	}
	if o.Limit <= 0 {
		o.Limit = 20
	}
	return o
}

type ListCommitsInput struct {
	Options  *ListOptions `json:"options"`
	Branch   string       `json:"branch"`
	FileName string       `json:"fileName"`
}

type CommitInfo struct {
	Commit *Commit `json:"commit"`
}

type ListCommitsOutput ListResult[*CommitInfo]

type ListBranchesInput struct {
	Options  *ListOptions `json:"options"`
	Remote   bool         `json:"remote"`
	WithHead bool         `json:"withHead"`
}

type BranchInfo struct {
	Ref    string  `json:"ref"`
	Name   string  `json:"name"`
	Commit *Commit `json:"commit"`
}

type ListBranchesOutput ListResult[*BranchInfo]

type FileNameData struct {
	Data    []byte `json:"data"` // NOTE: empty file is valid!
	Name    string `json:"name"`
	OldName string `json:"oldName"` // used when move/rename a file
}

type FileInfo struct {
	FileNameData
	Size      int64 `json:"size"`
	IsDir     bool  `json:"isDir"`
	IsBinary  bool  `json:"isBinary"`
	IsSymlink bool  `json:"isSymlink"`
}

type AddFilesInput struct {
	Branch    string          `json:"branch"`
	Files     []*FileNameData `json:"files"`
	Message   string          `json:"message"`
	Overwrite bool            `json:"overwrite"`
	Unpack    bool            `json:"unpack"`   // valid only if the given file is a tar gzip archive, the Files must only have one item
	Uploaded  bool            `json:"uploaded"` // indicates whether the files have been uploaded beforehand
}

type AddFilesOutput struct {
	Commit *Commit `json:"commit"`
}

// ListFilesInput list files under specified directory
type ListFilesInput struct {
	Branch string `json:"branch"`

	Dir string `json:"dir"`

	// WithFileContent only valid if file is a normal file
	WithFileContent bool `json:"withFileContent"`

	// WithLastCommit indicates whether retrieve the last commit for file
	WithLastCommit bool `json:"withLastCommit"`
}

type ListFilesOutput ListResult[*FileCommitInfo]

type DeleteFilesInput struct {
	Branch  string   `json:"branch"`
	Files   []string `json:"files"` // item can be a directory or file
	Message string   `json:"message"`
}

type DeleteFilesOutput struct {
	Commit *Commit `json:"commit"`
}

type CheckOutBranchInput struct {
	Branch string `json:"branch"`
	Force  bool   `json:"force"`
}

type CheckOutBranchOutput struct {
	WorkTree *git.Worktree `json:"-"`
}

type CommitAndPushInput struct {
	Branch   string        `json:"branch"`
	WorkTree *git.Worktree `json:"-"`
	Message  string        `json:"message"`
	SignOff  bool          `json:"signOff"`
}

type CommitAndPushOutput struct {
	Commit *Commit `json:"commit"`
}

type CleanAndPullInput struct {
	WorkTree *git.Worktree `json:"-"`
	Branch   string        `json:"branch"`
}

type CleanAndPullOutput struct {
}

type FileCommitInfo struct {
	FileInfo
	Commit *Commit `json:"commit"`
}

type DeleteCloneInput struct {
}

type DeleteCloneOutput struct {
}

type GetConfigInput struct {
}

type GetConfigOutput struct {
	Config *config.Config `json:"config"`
}

type UpdateConfigInput struct {
	Config *config.Config `json:"config"`
}

type UpdateConfigOutput struct {
	Config *config.Config `json:"config"`
}

type GetBranchInput struct {
	Branch string `json:"branch"`
	Remote bool   `json:"remote"`
}

type GetBranchOutput struct {
	Branch *BranchInfo `json:"branch"`
}

type GetCommitInput struct {
	Commit string `json:"commit"`
}

type GetCommitOutput struct {
	Commit *Commit `json:"commit"`
}

type GetFileInput struct {
	Commit string `json:"commit"`

	Branch string `json:"branch"` // only valid when Commit not present

	// File must be a regular file (including symbolic link to file), not a directory
	File string `json:"file"`

	WithFileContent bool `json:"withFileContent"`
}

type GetFileOutput struct {
	File   *FileInfo     `json:"file"`
	Reader io.ReadCloser `json:"-"` // Note: the reader might be closed if File.Data is not empty
}

type UploadFilesInput struct {
	Files []*FileNameData `json:"files"`
}

type UploadFilesOutput struct {
}

type GitRepoService interface {
	ListBranches(ctx context.Context, input *ListBranchesInput) (*ListBranchesOutput, error)
	GetBranch(ctx context.Context, input *GetBranchInput) (*GetBranchOutput, error)
	CheckOutBranch(ctx context.Context, input *CheckOutBranchInput) (*CheckOutBranchOutput, error)
	ListCommits(ctx context.Context, input *ListCommitsInput) (*ListCommitsOutput, error)
	GetCommit(ctx context.Context, input *GetCommitInput) (*GetCommitOutput, error)
	AddFiles(ctx context.Context, input *AddFilesInput) (*AddFilesOutput, error)
	UploadFiles(ctx context.Context, input *UploadFilesInput) (*UploadFilesOutput, error)
	DeleteFiles(ctx context.Context, input *DeleteFilesInput) (*DeleteFilesOutput, error)
	ListFiles(ctx context.Context, input *ListFilesInput) (*ListFilesOutput, error)
	GetFile(ctx context.Context, input *GetFileInput) (*GetFileOutput, error)
	CommitAndPush(ctx context.Context, input *CommitAndPushInput) (*CommitAndPushOutput, error)
	CleanAndPull(ctx context.Context, input *CleanAndPullInput) (*CleanAndPullOutput, error)
	DeleteClone(ctx context.Context, input *DeleteCloneInput) (*DeleteCloneOutput, error)
	GetConfig(ctx context.Context, input *GetConfigInput) (*GetConfigOutput, error)
	UpdateConfig(ctx context.Context, input *UpdateConfigInput) (*UpdateConfigOutput, error)
}

type Handler interface {
	ListBranches(req *restful.Request, res *restful.Response)
	GetBranch(req *restful.Request, res *restful.Response)
	CheckOutBranch(req *restful.Request, res *restful.Response)
	CleanAndPullBranch(req *restful.Request, res *restful.Response)
	ListCommits(req *restful.Request, res *restful.Response)
	GetCommit(req *restful.Request, res *restful.Response)
	AddFiles(req *restful.Request, res *restful.Response)
	UploadFiles(req *restful.Request, res *restful.Response)
	DeleteFiles(req *restful.Request, res *restful.Response)
	ListFiles(req *restful.Request, res *restful.Response)
	GetFile(req *restful.Request, res *restful.Response)
	DownloadFile(req *restful.Request, res *restful.Response)
	GetConfig(req *restful.Request, res *restful.Response)
	UpdateConfig(req *restful.Request, res *restful.Response)
}

type GitRepoFactory interface {
	NewRepoService(ctx context.Context, user user.Info, repo types.NamespacedName) (GitRepoService, error)
	DeleteRepoClone(ctx context.Context, repo types.NamespacedName) error
}

func ParseListOptionsFromRequest(req *restful.Request) *ListOptions {
	pageNumber, pageSize := common.GetPageParameters(req)
	listOptions := &ListOptions{
		Page:  pageNumber,
		Limit: pageSize,
	}
	return listOptions.Correct()
}

type GitRepoOptions struct {
	author *object.Signature
	user   user.Info
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
