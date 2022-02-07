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

package pipeline

import (
	"reflect"
	"testing"
)

func TestBranchSlice_SearchByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name       string
		branches   BranchSlice
		args       args
		wantExist  bool
		wantBranch *Branch
	}{{
		name: "Search for main",
		branches: BranchSlice{{
			Name: "main",
		}, {
			Name: "dev",
		}},
		args: args{
			name: "main",
		},
		wantExist: true,
		wantBranch: &Branch{
			Name: "main",
		},
	}, {
		name: "Search for dev",
		branches: BranchSlice{{
			Name: "main",
		}, {
			Name: "dev",
		}},
		args: args{
			name: "dev",
		},
		wantExist: true,
		wantBranch: &Branch{
			Name: "dev",
		},
	}, {
		name: "Search for nothing",
		branches: BranchSlice{{
			Name: "main",
		}, {
			Name: "dev",
		}},
		args: args{
			name: "nothing",
		},
		wantExist: false,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exist, branch := tt.branches.SearchByName(tt.args.name)
			if exist != tt.wantExist {
				t.Errorf("BranchSlice.SearchByName() got = %v, want %v", exist, tt.wantExist)
			}
			if !reflect.DeepEqual(branch, tt.wantBranch) {
				t.Errorf("BranchSlice.SearchByName() got1 = %v, want %v", branch, tt.wantBranch)
			}
		})
	}
}
