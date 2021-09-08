#!/bin/bash

# Copyright 2021 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

VERSION=$1

TKN_CRDS=("300-pipeline" "300-pipelinerun" "300-task")
 
for CRD in ${TKN_CRDS[@]}
do
  curl -L https://raw.githubusercontent.com/tektoncd/pipeline/$VERSION/config/$CRD.yaml -o controllers/tekton/testdata/$CRD.yaml
done
