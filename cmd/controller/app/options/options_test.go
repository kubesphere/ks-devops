/*
Copyright 2022 The KubeSphere Authors.

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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOption(t *testing.T) {
	opt := NewDevOpsControllerManagerOptions()
	assert.NotNil(t, opt)

	flags := opt.Flags()
	assert.NotNil(t, flags)
	assert.NotNil(t, flags.FlagSet("kubernetes"))
	assert.NotNil(t, flags.FlagSet("devops"))
	assert.NotNil(t, flags.FlagSet("feature"))
	assert.NotNil(t, flags.FlagSet("argocd"))
	assert.NotNil(t, flags.FlagSet("generic"))
	assert.NotNil(t, flags.FlagSet("leaderelection"))
	assert.NotNil(t, flags.FlagSet("klog"))

	opt.ApplicationSelector = "key=value"
	assert.Nil(t, opt.Validate())

	opt.ApplicationSelector = "!@#$"
	assert.NotNil(t, opt.Validate())
}
