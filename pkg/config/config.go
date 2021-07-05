/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"fmt"
	authoptions "kubesphere.io/devops/pkg/apiserver/authentication/options"
	"kubesphere.io/devops/pkg/client/cache"
	"kubesphere.io/devops/pkg/client/k8s"
	"kubesphere.io/devops/pkg/client/sonarqube"
	"reflect"
	"strings"

	"github.com/spf13/viper"

	"kubesphere.io/devops/pkg/client/devops/jenkins"
	"kubesphere.io/devops/pkg/client/s3"
)

// Package config saves configuration for running KubeSphere components
//
// Config can be configured from command line flags and configuration file.
// Command line flags hold higher priority than configuration file. But if
// component Endpoint/Host/APIServer was left empty, all of that component
// command line flags will be ignored, use configuration file instead.
// For example, we have configuration file
//
// mysql:
//   host: mysql.kubesphere-system.svc
//   username: root
//   password: password
//
// At the same time, have command line flags like following:
//
// --mysql-host mysql.openpitrix-system.svc --mysql-username king --mysql-password 1234
//
// We will use `king:1234@mysql.openpitrix-system.svc` from command line flags rather
// than `root:password@mysql.kubesphere-system.svc` from configuration file,
// cause command line has higher priority. But if command line flags like following:
//
// --mysql-username root --mysql-password password
//
// we will `root:password@mysql.kubesphere-system.svc` as input, cause
// mysql-host is missing in command line flags, all other mysql command line flags
// will be ignored.

const (
	// DefaultConfigurationName is the default name of configuration
	DefaultConfigurationName = "kubesphere"
	// DefaultConfigurationFileName is the default filename of configuration
	DefaultConfigurationFileName = DefaultConfigurationName + ".yaml"

	// DefaultConfigurationPath the default location of the configuration file
	defaultConfigurationPath = "/etc/kubesphere"
)

// AuthMode is the auth mode of current project
// TODO add a validation method to verify all the supported values
type AuthMode string

var (
	// AuthModeToken let it use the token directly
	AuthModeToken AuthMode = "token"
)

// Config defines everything needed for apiserver to deal with external services
type Config struct {
	DevopsOptions         *jenkins.Options                   `json:"devops,omitempty" yaml:"devops,omitempty" mapstructure:"devops"`
	KubernetesOptions     *k8s.KubernetesOptions             `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes"`
	RedisOptions          *cache.Options                     `json:"redis,omitempty" yaml:"redis,omitempty" mapstructure:"redis"`
	S3Options             *s3.Options                        `json:"s3,omitempty" yaml:"s3,omitempty" mapstructure:"s3"`
	SonarQubeOptions      *sonarqube.Options                 `json:"sonarqube,omitempty" yaml:"sonarQube,omitempty" mapstructure:"sonarqube"`
	AuthenticationOptions *authoptions.AuthenticationOptions `json:"authentication,omitempty" yaml:"authentication,omitempty" mapstructure:"authentication"`
	AuthMode              AuthMode                           `json:"authMode,omitempty" yaml:"authMode,omitempty" mapstructure:"authMode"`
	JWTSecret             string                             `json:"jwtSecret,omitempty" yaml:"jwtSecret,omitempty" mapstructure:"jwtSecret"`
}

// newConfig creates a default non-empty Config
func New() *Config {
	return &Config{
		SonarQubeOptions:  sonarqube.NewSonarQubeOptions(),
		DevopsOptions:     jenkins.NewDevopsOptions(),
		KubernetesOptions: k8s.NewKubernetesOptions(),
		S3Options:         s3.NewS3Options(),
		AuthMode:          AuthModeToken,
	}
}

// TryLoadFromDisk loads configuration from default location after server startup
// return nil error if configuration file not exists
func TryLoadFromDisk() (*Config, error) {
	viper.SetConfigName(DefaultConfigurationName)
	viper.AddConfigPath(defaultConfigurationPath)

	// Load from current working directory, only used for debugging
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, err
		} else {
			return nil, fmt.Errorf("error parsing configuration file %s", err)
		}
	}

	conf := New()

	if err := viper.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

// convertToMap simply converts config to map[string]bool
// to hide sensitive information
func (conf *Config) ToMap() map[string]bool {
	conf.stripEmptyOptions()
	result := make(map[string]bool, 0)

	if conf == nil {
		return result
	}

	c := reflect.Indirect(reflect.ValueOf(conf))

	for i := 0; i < c.NumField(); i++ {
		name := strings.Split(c.Type().Field(i).Tag.Get("json"), ",")[0]
		if strings.HasPrefix(name, "-") {
			continue
		}

		if c.Field(i).IsNil() {
			result[name] = false
		} else {
			result[name] = true
		}
	}

	return result
}

// Remove invalid options before serializing to json or yaml
func (conf *Config) stripEmptyOptions() {
	if conf.DevopsOptions != nil && conf.DevopsOptions.Host == "" {
		conf.DevopsOptions = nil
	}

	if conf.S3Options != nil && conf.S3Options.Endpoint == "" {
		conf.S3Options = nil
	}
}
