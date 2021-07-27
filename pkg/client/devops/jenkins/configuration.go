package jenkins

import (
	"errors"
	"net/http"
)

// According to: https://github.com/jenkinsci/configuration-as-code-plugin/blob/master/docs/features/configurationReload.md

const reloadEndpoint = "/configuration-as-code/reload/"

func (jenkins *Jenkins) ReloadConfiguration() error {
	// reload Jenkins CasC
	response, err := jenkins.Requester.Post(reloadEndpoint, nil, nil, nil)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusFound {
		return errors.New("failed to reload Jenkins CasC")
	}
	// reload successfully
	return nil
}
