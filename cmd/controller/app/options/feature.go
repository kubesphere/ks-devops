/*
Copyright 2021 KubeSphere Authors

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

package options

import (
	"strings"

	"github.com/spf13/pflag"
	cliflag "k8s.io/component-base/cli/flag"
	"kubesphere.io/devops/pkg/utils/reflectutils"
)

// FeatureOptions provide some feature options, such as specifying the controller to be enabled.
type FeatureOptions struct {
	Controllers map[string]bool
}

// GetControllers returns the controllers map
// it supports a special key 'all', the default config is not working if 'all' is false
func (o *FeatureOptions) GetControllers() map[string]bool {
	defaultMap := map[string]bool{
		"jenkins":       true,
		"jenkinsconfig": true,
	}

	// support to only enable the specific controllers
	if val, ok := o.Controllers["all"]; ok {
		delete(o.Controllers, "all")
		if !val {
			return o.Controllers
		}
	}

	// merge the default values and input from users
	for key, val := range o.Controllers {
		defaultMap[key] = val
	}
	return defaultMap
}

// NewFeatureOptions provide default options
func NewFeatureOptions() *FeatureOptions {
	return &FeatureOptions{}
}

// Validate checks validation of FeatureOptions.
func (o *FeatureOptions) Validate() []error {
	return []error{}
}

// ApplyTo fills up FeatureOptions config with options
func (o *FeatureOptions) ApplyTo(options *FeatureOptions) {
	reflectutils.Override(options, o)
}

// AddFlags adds flags related to FeatureOptions for controller manager to the feature FlagSet.
func (o *FeatureOptions) AddFlags(fs *pflag.FlagSet, c *FeatureOptions) {
	fs.Var(cliflag.NewMapStringBool(&o.Controllers), "enabled-controllers", "A set of key=value pairs that describe feature options for controllers. "+
		"Options are:\n"+strings.Join(c.knownControllers(), "\n"))
}

func (o *FeatureOptions) knownControllers() []string {
	controllers := make([]string, 0, len(o.Controllers))
	for name := range o.Controllers {
		controllers = append(controllers, name)
	}
	return controllers
}
