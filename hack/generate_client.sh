#!/bin/bash

set -e

GV="$1"

./hack/generate_group.sh all kubesphere.io/devops/pkg/client kubesphere.io/devops/api "${GV}" --output-base=./  -h "$PWD/hack/boilerplate.go.txt"
