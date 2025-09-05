package gitops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/kubesphere/ks-devops/pkg/api/devops/v1alpha3"
	"github.com/kubesphere/ks-devops/pkg/config"
	"github.com/kubesphere/ks-devops/pkg/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/authentication/user"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type gitRepoFactory struct {
	k8sClient client.Client
	config    *config.GitOpsOptions
}

func (g *gitRepoFactory) DeleteRepoClone(ctx context.Context, repoName types.NamespacedName) error {
	gitRepo, err := g.getAndCheckGitRepo(ctx, repoName)
	if err != nil {
		return err
	}

	repoDir := g.getRepoDirForUser(ctx, repoName.Namespace, gitRepo.Spec.URL)

	err = os.RemoveAll(repoDir)
	return err
}

func (g *gitRepoFactory) parseSkipTLSFromRepo(ctx context.Context, repo *v1alpha3.GitRepository) bool {
	v, ok := repo.Annotations[constants.InsecureSkipTLSAnnotationKey]
	if !ok {
		return false
	}
	yes, _ := strconv.ParseBool(v)
	return yes
}

func (g *gitRepoFactory) parseTLSCertsFromRepo(ctx context.Context, repo *v1alpha3.GitRepository) ([]byte, error) {
	tlsCertsName := repo.Annotations[constants.TLSCertsNameAnnotationKey]
	if tlsCertsName == "" {
		return nil, nil
	}

	tlsCertsNamespace := repo.Annotations[constants.TLSCertsNameSpaceAnnotationKey]
	if tlsCertsNamespace == "" {
		tlsCertsNamespace = constants.DevOpsWorkerNamespace
	}

	caSecret := &v1.ConfigMap{}
	if err := g.k8sClient.Get(ctx, types.NamespacedName{
		Name:      tlsCertsName,
		Namespace: tlsCertsNamespace,
	}, caSecret); err != nil {
		return nil, err
	}

	if caSecret.Data[constants.TLSCertKey] == "" {
		return nil, fmt.Errorf("invalid secret data: %s", constants.TLSCertKey)
	}

	return []byte(caSecret.Data[constants.TLSCertKey]), nil
}

func (g *gitRepoFactory) parseAuthorFromSecret(ctx context.Context, secret *v1.Secret) *object.Signature {
	var authorName, authorEmail string
	if secret != nil {
		authorName = secret.Annotations[constants.GitAuthorNameAnnotationKey]
		authorEmail = secret.Annotations[constants.GitAuthorEmailAnnotationKey]
	}
	return &object.Signature{
		Name:  authorName,
		Email: authorEmail,
	}
}

func (g *gitRepoFactory) NewRepoService(ctx context.Context, user user.Info, repoName types.NamespacedName) (GitRepoService, error) {
	gitRepo, err := g.getAndCheckGitRepo(ctx, repoName)
	if err != nil {
		return nil, err
	}

	var token, tokenUser string
	var secret *v1.Secret
	var auth *http.BasicAuth

	secretRef := gitRepo.Spec.Secret
	// Note: public repo does not need a secret
	if secretRef != nil && secretRef.Name != "" && secretRef.Namespace != "" {
		secret, err = g.getSecret(ctx, secretRef)
		if err != nil {
			return nil, err
		}
		token, tokenUser, err = g.getTokenFromSecret(ctx, secret)
		if err != nil {
			return nil, err
		}
		if token == "" {
			return nil, fmt.Errorf("failed to get token")
		}
	}

	if tokenUser == "" {
		tokenUser = user.GetName() // yes, this can be anything except an empty string
	}

	if tokenUser != "" && token != "" {
		auth = &http.BasicAuth{
			Username: tokenUser,
			Password: token,
		}
	}

	insecureSkipTLS := g.parseSkipTLSFromRepo(ctx, gitRepo)
	ca, err := g.parseTLSCertsFromRepo(ctx, gitRepo)
	if err != nil {
		return nil, err
	}
	author := g.parseAuthorFromSecret(ctx, secret)
	repoDir := g.getRepoDirForUser(ctx, repoName.Namespace, gitRepo.Spec.URL)

	if author.Name == "" {
		author.Name = tokenUser
	}

	var repo *git.Repository
	_, err = os.Stat(repoDir)
	if err == nil {
		repo, err = git.PlainOpen(repoDir)
		if err != nil {
			return nil, err
		}
	} else if os.IsNotExist(err) {
		opt := &git.CloneOptions{
			Auth:            auth,
			URL:             gitRepo.Spec.URL,
			Progress:        os.Stdout,
			InsecureSkipTLS: insecureSkipTLS,
			CABundle:        ca,
		}
		repo, err = git.PlainClone(repoDir, false, opt)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	if repo == nil {
		return nil, fmt.Errorf("failed to create or open repository: %s", repoDir)
	}

	gitRepoOpts := &GitRepoOptions{
		author:          author,
		user:            user,
		repo:            repo,
		auth:            auth,
		newFilePerm:     g.config.NewFilePerm,
		insecureSkipTLS: insecureSkipTLS,
		caBundle:        ca,
	}
	repoService := NewGitRepoService(gitRepoOpts)

	return repoService, nil
}

func (g *gitRepoFactory) getRepoDirForUser(ctx context.Context, namespace string, repoURL string) string {
	repoURL = strings.TrimPrefix(repoURL, "http://")
	repoURL = strings.TrimPrefix(repoURL, "https://")
	repoURL = strings.TrimPrefix(repoURL, "git@")
	repoURL = strings.Replace(repoURL, ":", "/", 1)
	root := g.config.RootDir
	if root == "" {
		root = "/gitops"
	}
	dir := filepath.Join(root, namespace, repoURL)
	return dir
}

func (g *gitRepoFactory) getAndCheckGitRepo(ctx context.Context, repoName types.NamespacedName) (*v1alpha3.GitRepository, error) {
	gitRepo := &v1alpha3.GitRepository{}
	err := g.k8sClient.Get(ctx, repoName, gitRepo)
	if err != nil {
		return nil, err
	}

	if gitRepo.Spec.URL == "" {
		return nil, fmt.Errorf("git repository URL is not specified")
	}

	if gitRepo.Spec.Secret == nil {
		return nil, fmt.Errorf("git repository secret is not specified")
	}

	return gitRepo, nil
}

func (g *gitRepoFactory) getTokenFromSecret(ctx context.Context, gitSecret *v1.Secret) (token, username string, err error) {
	switch gitSecret.Type {
	case v1.SecretTypeBasicAuth, v1alpha3.SecretTypeBasicAuth:
		token = string(gitSecret.Data[v1.BasicAuthPasswordKey])
		username = string(gitSecret.Data[v1.BasicAuthUsernameKey])
	case v1.SecretTypeOpaque:
		token = string(gitSecret.Data[v1.ServiceAccountTokenKey])
	case v1alpha3.SecretTypeSecretText:
		token = string(gitSecret.Data["secret"])
	}

	return token, username, err
}

func (g *gitRepoFactory) getSecret(ctx context.Context, ref *v1.SecretReference) (*v1.Secret, error) {
	secret := &v1.Secret{}
	err := g.k8sClient.Get(ctx, types.NamespacedName{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}, secret)
	return secret, err
}

var _ GitRepoFactory = &gitRepoFactory{}

func NewGitRepoFactory(k8sClient client.Client, config *config.GitOpsOptions) GitRepoFactory {
	return &gitRepoFactory{
		k8sClient: k8sClient,
		config:    config,
	}
}
