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

package internal

import "testing"

func TestCommonSituation(t *testing.T) {
	// make sure these functions do not panic
	// I add these test cases because it's possible that users just do give the git source
	AppendGitlabSourceToEtree(nil, nil)
	AppendGithubSourceToEtree(nil, nil)
	AppendBitbucketServerSourceToEtree(nil, nil)
	AppendGitSourceToEtree(nil, nil)
	AppendSingleSvnSourceToEtree(nil, nil)
	AppendSvnSourceToEtree(nil, nil)
}
