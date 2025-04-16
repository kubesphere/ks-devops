package scm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func getAuthMethod(repoURL, username, password string, sshKey []byte) (transport.AuthMethod, error) {
	switch {
	case strings.HasPrefix(repoURL, "http://"), strings.HasPrefix(repoURL, "https://"):
		if password == "" {
			return nil, errors.New("password/token required for HTTP URLs")
		}
		return &http.BasicAuth{Username: username, Password: password}, nil

	case strings.HasPrefix(repoURL, "git@"):
		fallthrough
	case strings.Contains(repoURL, "ssh://"):
		if len(sshKey) == 0 {
			return nil, errors.New("SSH private key required for SSH URLs")
		}

		publicKeys, err := ssh.NewPublicKeys(username, sshKey, password)
		if err != nil {
			return nil, fmt.Errorf("failed to create SSH auth: %w", err)
		}
		publicKeys.HostKeyCallback = gossh.InsecureIgnoreHostKey()

		return publicKeys, nil

	default:
		return nil, errors.New("unsupported repository URL scheme")
	}
}

func handleGitError(err error) error {
	switch {
	case errors.Is(err, transport.ErrAuthenticationRequired):
		return fmt.Errorf("authentication failed: %v", err)
	case errors.Is(err, transport.ErrAuthorizationFailed):
		return fmt.Errorf("authorization failed: %v", err)
	case errors.Is(err, transport.ErrRepositoryNotFound):
		return fmt.Errorf("repository not found: %v", err)
	case strings.Contains(err.Error(), "connection refused"),
		strings.Contains(err.Error(), "no such host"),
		strings.Contains(err.Error(), "timed out"):
		return fmt.Errorf("connectivity error: %v", err)
	default:
		return fmt.Errorf("access verification failed: %v", err)
	}
}
