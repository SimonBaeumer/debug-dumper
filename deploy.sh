#!/usr/bin/env bash
set -eo pipefail

kubectl kustomize ./metrics-server | kubectl apply -f -