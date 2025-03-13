#!/bin/bash

set -e

GV="$1"

./hack/generate_group.sh all github.com/kubesphere/ks-devops/pkg/client github.com/kubesphere/ks-devops/api "${GV}" --output-base=./  -h "$PWD/hack/boilerplate.go.txt"
