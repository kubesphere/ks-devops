package gitops

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/emicklei/go-restful/v3"
	apiserverRequest "github.com/kubesphere/ks-devops/pkg/apiserver/request"
	"github.com/kubesphere/ks-devops/pkg/config"
	"github.com/kubesphere/ks-devops/pkg/kapis"
	"github.com/kubesphere/ks-devops/pkg/kapis/common"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var DefaultGitRepoFactory GitRepoFactory

type handler struct {
	k8sClient client.Client
	config    *config.GitOpsOptions
	factory   GitRepoFactory
}

func (h *handler) DownloadFile(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	branch := common.GetPathParameter(req, pathParameterBranch)
	commit := common.GetPathParameter(req, pathParameterCommit)
	fileEncoded := common.GetPathParameter(req, pathParameterFile)
	file, err := base64.StdEncoding.DecodeString(fileEncoded)
	if err != nil {
		kapis.HandleBadRequest(res, req, err)
		return
	}

	getFileOut, err := repoService.GetFile(ctx, &GetFileInput{
		Branch:          branch,
		Commit:          commit,
		File:            string(file),
		WithFileContent: true,
	})
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	var bs []byte
	if len(getFileOut.File.Data) > 0 {
		bs = getFileOut.File.Data
	} else if getFileOut.Reader != nil {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(getFileOut.Reader)
		defer getFileOut.Reader.Close()
		if err != nil {
			kapis.HandleInternalError(res, req, err)
			return
		}
		bs = buf.Bytes()
	} else {
		kapis.HandleInternalError(res, req, errors.New("no data to read"))
		return
	}

	ext := filepath.Ext(getFileOut.File.Name)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	res.AddHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", getFileOut.File.Name))
	res.Header().Set(restful.HEADER_ContentType, mimeType)
	res.WriteHeader(http.StatusOK)

	_, err = res.ResponseWriter.Write(bs)
	if err != nil {
		kapis.HandleInternalError(res, req, err)
		return
	}
}

func (h *handler) GetFile(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	branch := common.GetPathParameter(req, pathParameterBranch)
	commit := common.GetPathParameter(req, pathParameterCommit)
	fileEncoded := common.GetPathParameter(req, pathParameterFile)
	file, err := base64.StdEncoding.DecodeString(fileEncoded)
	if err != nil {
		kapis.HandleBadRequest(res, req, err)
		return
	}
	withContent, _ := strconv.ParseBool(common.GetQueryParameter(req, queryParameterWithContent))

	getFileOut, err := repoService.GetFile(ctx, &GetFileInput{
		Branch:          branch,
		Commit:          commit,
		File:            string(file),
		WithFileContent: withContent,
	})
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}
	out := getFileOut.File
	_ = res.WriteEntity(out)
}

func (h *handler) UploadFiles(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	var files []*FileNameData
	for i := 0; i < 10; i++ {
		filename := fmt.Sprintf("file%d", i)
		file, header, err := req.Request.FormFile(filename)
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				break
			} else {
				kapis.HandleError(req, res, err)
				return
			}
		}
		if header.Size > UploadDownloadFileSizeLimit {
			kapis.HandleBadRequest(res, req, fmt.Errorf("file %s size exceeds limit %d bytes", filename, UploadDownloadFileSizeLimit))
			return
		}

		data, err := io.ReadAll(file)
		if err != nil {
			kapis.HandleError(req, res, err)
			return
		}

		name := req.Request.FormValue(fmt.Sprintf("%s_name", filename))
		if len(name) == 0 {
			kapis.HandleBadRequest(res, req, fmt.Errorf("no name provided for file %s", filename))
			return
		}

		fi := &FileNameData{
			Data: data,
			Name: name,
		}
		files = append(files, fi)
	}

	if len(files) == 0 {
		kapis.HandleBadRequest(res, req, errors.New("no files provided"))
		return
	}

	uploadFilesInput := &UploadFilesInput{
		Files: files,
	}

	out, err := repoService.UploadFiles(ctx, uploadFilesInput)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}
	_ = res.WriteEntity(out)
}

func (h *handler) GetBranch(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	branch := common.GetPathParameter(req, pathParameterBranch)

	out, err := repoService.GetBranch(ctx, &GetBranchInput{
		Branch: branch,
		Remote: true,
	})

	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	_ = res.WriteEntity(out.Branch)
}

func (h *handler) GetCommit(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	out, err := repoService.GetCommit(ctx, &GetCommitInput{
		Commit: common.GetPathParameter(req, pathParameterCommit),
	})

	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	_ = res.WriteEntity(out.Commit)
}

func (h *handler) GetConfig(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}
	out, err := repoService.GetConfig(ctx, &GetConfigInput{})
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}
	_ = res.WriteEntity(out)
}

func (h *handler) UpdateConfig(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	input := &UpdateConfigInput{}
	err := req.ReadEntity(input)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	out, err := repoService.UpdateConfig(ctx, input)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}
	_ = res.WriteEntity(out)
}

func (h *handler) getRepoService(req *restful.Request) (GitRepoService, error) {
	ctx := req.Request.Context()
	ns := common.GetPathParameter(req, common.NamespacePathParameter)
	repo := common.GetPathParameter(req, pathParameterGitRepository)

	requestUser, ok := apiserverRequest.UserFrom(req.Request.Context())
	if !ok {
		err := fmt.Errorf("cannot obtain user info")
		klog.Errorln(err)
		return nil, err
	}

	repoService, err := h.factory.NewRepoService(ctx, requestUser, types.NamespacedName{
		Namespace: ns,
		Name:      repo,
	})

	return repoService, err
}

func (h *handler) ListBranches(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	withHeadStr := common.GetQueryParameter(req, queryParameterWithHead)
	withHead, _ := strconv.ParseBool(withHeadStr)
	listOptions := ParseListOptionsFromRequest(req)
	listBranchesInput := &ListBranchesInput{
		Options:  listOptions,
		Remote:   true,
		WithHead: withHead,
	}
	listBranchesOutput, err := repoService.ListBranches(ctx, listBranchesInput)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}
	_ = res.WriteEntity(listBranchesOutput)
}

func (h *handler) CheckOutBranch(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	branch := common.GetPathParameter(req, pathParameterBranch)
	out, err := repoService.CheckOutBranch(ctx, &CheckOutBranchInput{
		Branch: branch,
		Force:  true,
	})
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}
	_ = res.WriteEntity(out)
}

func (h *handler) CleanAndPullBranch(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	branch := common.GetPathParameter(req, pathParameterBranch)

	coOut, err := repoService.CheckOutBranch(ctx, &CheckOutBranchInput{
		Branch: branch,
		Force:  false,
	})
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	out, err := repoService.CleanAndPull(ctx, &CleanAndPullInput{
		WorkTree: coOut.WorkTree,
		Branch:   branch,
	})
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}
	_ = res.WriteEntity(out)
}

func (h *handler) ListCommits(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	listOptions := ParseListOptionsFromRequest(req)
	branch := common.GetPathParameter(req, pathParameterBranch)
	fileName := common.GetQueryParameter(req, queryParameterFile)

	out, err := repoService.ListCommits(ctx, &ListCommitsInput{
		Options:  listOptions,
		Branch:   branch,
		FileName: fileName,
	})
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}
	_ = res.WriteEntity(out)
}

func (h *handler) AddFiles(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	branch := common.GetPathParameter(req, pathParameterBranch)

	addFilesInput := &AddFilesInput{}
	err = req.ReadEntity(addFilesInput)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}
	addFilesInput.Branch = branch

	out, err := repoService.AddFiles(ctx, addFilesInput)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}
	_ = res.WriteEntity(out)
}

func (h *handler) DeleteFiles(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	branch := common.GetPathParameter(req, pathParameterBranch)
	files := common.GetQueryParameters(req, queryParameterFile)
	message := common.GetQueryParameter(req, queryParameterMessage)

	input := &DeleteFilesInput{
		Branch:  branch,
		Files:   files,
		Message: message,
	}

	out, err := repoService.DeleteFiles(ctx, input)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}
	_ = res.WriteEntity(out)
}

func (h *handler) ListFiles(req *restful.Request, res *restful.Response) {
	ctx := req.Request.Context()
	repoService, err := h.getRepoService(req)
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}

	branch := common.GetPathParameter(req, pathParameterBranch)
	file := common.GetQueryParameter(req, queryParameterFile)
	withContent, _ := strconv.ParseBool(common.GetQueryParameter(req, queryParameterWithContent))
	withLastCommit, _ := strconv.ParseBool(common.GetQueryParameter(req, queryParameterWithLastCommit))

	out, err := repoService.ListFiles(ctx, &ListFilesInput{
		Branch:          branch,
		Dir:             file,
		WithFileContent: withContent,
		WithLastCommit:  withLastCommit,
	})
	if err != nil {
		kapis.HandleError(req, res, err)
		return
	}
	_ = res.WriteEntity(out)
}

var _ Handler = &handler{}

func NewHandler(k8sClient client.Client, config *config.GitOpsOptions) Handler {
	DefaultGitRepoFactory = NewGitRepoFactory(k8sClient, config)
	return &handler{
		k8sClient: k8sClient,
		config:    config,
		factory:   DefaultGitRepoFactory,
	}
}
