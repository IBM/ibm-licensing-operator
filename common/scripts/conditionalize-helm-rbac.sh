#!/usr/bin/env bash

# Copyright 2026 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Injects conditional RBAC Helm guards into generated RBAC templates, wrapping
# designated rule blocks / resource list items in
#   {{- if eq (include "<helper>" .) "true" }} ... {{- end }}
# guards driven by common/makefile-generate/helm-rbac-gating-table.
#
# Usage: conditionalize-helm-rbac.sh <dir> [--check]
#   <dir>      directory holding the RBAC YAML files to process
#   --check    do not modify files; exit 1 (printing a diff) if any target file
#              is missing or has stale guards (drift detection for CI)
#
# The script is idempotent (see the .awk for details), so --check on
# already-guarded templates is a no-op.

set -euo pipefail

DIR="${1:?usage: conditionalize-helm-rbac.sh <dir> [--check]}"
MODE="${2:-apply}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AWK_PROG="${SCRIPT_DIR}/conditionalize-helm-rbac.awk"
TABLE="${SCRIPT_DIR}/../makefile-generate/helm-rbac-gating-table"

rc=0

process() {
  local name="$1"
  local f="${DIR}/${name}"
  [ -f "$f" ] || return 0
  awk -v TABLE="$TABLE" -v FILEBN="$name" -f "$AWK_PROG" "$f" > "${f}.tmp"
  if [ "$MODE" = "--check" ]; then
    if ! diff -u "$f" "${f}.tmp"; then
      echo "DRIFT: ${name} is missing or has stale conditional RBAC guards." >&2
      echo "       Re-run: make generate-yaml-argo-cd (or conditionalize-helm-rbac.sh)." >&2
      rc=1
    fi
    rm -f "${f}.tmp"
  else
    mv "${f}.tmp" "$f"
  fi
}

process cluster-rbac.yaml
process rbac.yaml
process cluster-rbac-for-operandrequests.yaml

exit "$rc"
