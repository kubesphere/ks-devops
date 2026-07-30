#!/usr/bin/env bash

# Copyright 2026 The KubeSphere Authors.
# Licensed under the Apache License, Version 2.0.

set -euo pipefail

REPOSITORY_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SAMPLE_DIRECTORY="${REPOSITORY_ROOT}/config/samples/tekton"
DEMO_NAMESPACE="${DEMO_NAMESPACE:-kse-tekton-demo}"

command -v kubectl >/dev/null

kubectl get crd pipelines.tekton.dev >/dev/null
kubectl get crd pipelineruns.tekton.dev >/dev/null
kubectl get crd pipelines.devops.kubesphere.io >/dev/null
kubectl get crd pipelineruns.devops.kubesphere.io >/dev/null

kubectl apply -k "${SAMPLE_DIRECTORY}"

KSE_RUN_RESOURCE="$(kubectl create -f "${SAMPLE_DIRECTORY}/kse-pipelinerun.yaml" -o name)"
KSE_RUN_NAME="${KSE_RUN_RESOURCE##*/}"

echo "Created KubeSphere PipelineRun ${DEMO_NAMESPACE}/${KSE_RUN_NAME}"
kubectl wait --namespace "${DEMO_NAMESPACE}" \
  --for=condition=Succeeded=True \
  "pipelinerun.tekton.dev/${KSE_RUN_NAME}" \
  --timeout=10m
kubectl wait --namespace "${DEMO_NAMESPACE}" \
  --for=jsonpath='{.status.phase}'=Succeeded \
  "pipelinerun.devops.kubesphere.io/${KSE_RUN_NAME}" \
  --timeout=30s

kubectl get --namespace "${DEMO_NAMESPACE}" \
  "pipelinerun.devops.kubesphere.io/${KSE_RUN_NAME}" \
  -o custom-columns=NAME:.metadata.name,ENGINE:.spec.pipelineSpec.engine.type,NATIVE:.metadata.annotations.devops\\.kubesphere\\.io/tekton-pipelinerun,PHASE:.status.phase
kubectl get --namespace "${DEMO_NAMESPACE}" \
  "pipelinerun.tekton.dev/${KSE_RUN_NAME}"
