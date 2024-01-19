#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

source "${CODEGEN_PKG}/kube_codegen.sh"

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.

kube::codegen::gen_helpers \
    --input-pkg-root github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis \
    --output-base "$(dirname "${BASH_SOURCE[0]}")/.." \
    --boilerplate "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt"

#kube::codegen::gen_openapi \
#    --input-pkg-root github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis \
#    --output-pkg-root github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apiserver \
#    --output-base "$(dirname "${BASH_SOURCE[0]}")/.." \
#    --boilerplate "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt"

kube::codegen::gen_client \
    --with-watch \
    --input-pkg-root github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis \
    --output-pkg-root github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/clients \
    --output-base "$(dirname "${BASH_SOURCE[0]}")/.." \
    --boilerplate "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt"

