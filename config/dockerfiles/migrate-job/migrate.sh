#!/bin/sh

kubectl -n kubesphere-system get cm kubesphere-config -ojsonpath={.data.kubesphere\\.yaml} | \
  yq e '.devops.enable=false' - | yq e '.devops.devopsServiceAddress="127.0.0.1:9091"' -

export JENKINS_TOKEN=$(kubectl -n kubesphere-system get cm kubesphere-config -ojsonpath={.data.kubesphere\\.yaml} | \
  yq e '.devops.password' -)

kubectl -n kubesphere-devops-system get cm devops-config -ojsonpath={.data.kubesphere\\.yaml} | \
  yq e '.devops.password = strenv(JENKINS_TOKEN)' -
