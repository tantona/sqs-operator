#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

vendor/k8s.io/code-generator/generate-groups.sh \
deepcopy \
github.com/tantona/sqs-queue-operator/pkg/generated \
github.com/tantona/sqs-queue-operator/pkg/apis \
stable:v1 \
--go-header-file "./tmp/codegen/boilerplate.go.txt"
