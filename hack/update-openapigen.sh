#!/bin/bash
set -eux -o pipefail

go install k8s.io/kube-openapi/cmd/openapi-gen

openapi-gen \
  --go-header-file ./hack/custom-boilerplate.go.txt \
  --input-dirs ./pkg/apis/k8s.cni.cncf.io/v1 \
  --input-dirs ./vendor/k8s.io/apimachinery/pkg/apis/meta/v1 \
  --output-base pkg \
  --output-package apis/k8s.cni.cncf.io/v1 \
  --output-file-base openapi_generated \
  $@

