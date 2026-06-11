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

# Tests for conditionalize-helm-rbac.sh (ILS-2352).
#
# Self-contained POSIX-ish bash (no bats dependency). Covers:
#   1. fixture in/out      -- un-guarded fixtures -> byte-exact golden
#   2. idempotency         -- re-run / --check is a no-op
#   3. context scoping     -- operator role's own namespaces[get] stays unguarded
#   4. empty-guard         -- restricted role gets the outer "or" wrapper
#   5. missing file        -- absent opreqs file is skipped, exit 0
#   6. render matrix        -- helm template add/remove behavior (skipped if no helm)
#
# Usage: bash common/scripts/tests/conditionalize-helm-rbac-test.sh

set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPTS="${HERE}/.."
FIX="${HERE}/fixtures"
SCRIPT="${SCRIPTS}/conditionalize-helm-rbac.sh"
CHART="${HERE}/../../../deploy/argo-cd/components/license-service/helm-cluster-scoped"

pass=0
fail=0
ok()   { echo "ok   - $1"; pass=$((pass + 1)); }
bad()  { echo "FAIL - $1"; fail=$((fail + 1)); }

# ---------------------------------------------------------------------------
# 1. fixture in/out: un-guarded input -> byte-exact golden
# ---------------------------------------------------------------------------
wd="$(mktemp -d)"
cp "${FIX}/cluster-rbac.input.yaml"                   "${wd}/cluster-rbac.yaml"
cp "${FIX}/rbac.input.yaml"                           "${wd}/rbac.yaml"
cp "${FIX}/cluster-rbac-for-operandrequests.input.yaml" "${wd}/cluster-rbac-for-operandrequests.yaml"

bash "$SCRIPT" "$wd" >/dev/null 2>&1

for f in cluster-rbac.yaml rbac.yaml cluster-rbac-for-operandrequests.yaml; do
  golden="${FIX}/${f%.yaml}.golden.yaml"
  if diff -u "$golden" "${wd}/${f}" >/dev/null; then
    ok "fixture in/out: ${f} matches golden"
  else
    bad "fixture in/out: ${f} differs from golden"
    diff -u "$golden" "${wd}/${f}" | sed 's/^/      /'
  fi
done

# ---------------------------------------------------------------------------
# 2. idempotency: a second apply changes nothing; --check passes
# ---------------------------------------------------------------------------
before="$(cat "${wd}/cluster-rbac.yaml" "${wd}/rbac.yaml" "${wd}/cluster-rbac-for-operandrequests.yaml")"
bash "$SCRIPT" "$wd" >/dev/null 2>&1
after="$(cat "${wd}/cluster-rbac.yaml" "${wd}/rbac.yaml" "${wd}/cluster-rbac-for-operandrequests.yaml")"
if [ "$before" = "$after" ]; then ok "idempotency: second apply is a no-op"; else bad "idempotency: second apply changed output"; fi

if bash "$SCRIPT" "$wd" --check >/dev/null 2>&1; then
  ok "idempotency: --check passes on guarded files"
else
  bad "idempotency: --check reported drift on freshly guarded files"
fi

# ---------------------------------------------------------------------------
# 3. context scoping: the namespaceDiscovery guard wraps only the two operand
#    cluster roles, NOT the ibm-licensing-operator role's own namespaces[get].
# ---------------------------------------------------------------------------
# The operator role's namespaces rule must remain unconditional.
# Extract the ibm-licensing-operator ClusterRole block and assert its
# "- namespaces" line is not preceded by a namespaceDiscovery guard.
op_block="$(awk '/name: ibm-licensing-operator$/{f=1} f&&/^---/{if(seen)f=0; seen=1} f' "${wd}/cluster-rbac.yaml")"
if printf '%s\n' "$op_block" | grep -B1 '^      - namespaces$' | grep -q 'namespaceDiscoveryEnabled'; then
  bad "context scoping: operator role's namespaces[get] was wrongly guarded"
else
  ok "context scoping: operator role's namespaces[get] stays unconditional"
fi

# ---------------------------------------------------------------------------
# 4. empty-guard: restricted role gets the outer 'or' wrapper around its
#    fully-gated namespaces+nodes rule.
# ---------------------------------------------------------------------------
if grep -q '{{- if or (eq (include "ibm-licensing.namespaceDiscoveryEnabled" .) "true") (eq (include "ibm-licensing.nodeCpuCappingEnabled" .) "true") }}' "${wd}/cluster-rbac.yaml"; then
  ok "empty-guard: restricted role has outer 'or' wrapper"
else
  bad "empty-guard: missing outer 'or' wrapper on restricted role"
fi

rm -rf "$wd"

# ---------------------------------------------------------------------------
# 5. missing file: opreqs file absent -> skipped, exit 0, not created
# ---------------------------------------------------------------------------
wd2="$(mktemp -d)"
cp "${FIX}/rbac.input.yaml" "${wd2}/rbac.yaml"
if bash "$SCRIPT" "$wd2" >/dev/null 2>&1 \
    && [ ! -f "${wd2}/cluster-rbac.yaml" ] \
    && [ ! -f "${wd2}/cluster-rbac-for-operandrequests.yaml" ]; then
  ok "missing file: absent targets skipped, exit 0"
else
  bad "missing file: script errored or created absent targets"
fi
rm -rf "$wd2"

# ---------------------------------------------------------------------------
# 6. render matrix: prove helpers + guards compose under `helm template`.
# ---------------------------------------------------------------------------
if ! command -v helm >/dev/null 2>&1; then
  echo "skip - render matrix (helm not installed)"
else
  render() { helm template t "$CHART" "$@" 2>/dev/null; }
  # Patterns are EREs (anchored where it matters) so none start with '-'.
  has()    { printf '%s\n' "$2" | grep -Eq "$1"; }
  assert_present() { if has "$1" "$3"; then ok "render[$2]: present /$1/"; else bad "render[$2]: MISSING /$1/"; fi; }
  assert_absent()  { if has "$1" "$3"; then bad "render[$2]: UNEXPECTED /$1/"; else ok "render[$2]: absent /$1/"; fi; }

  RES_NODES='^ +- nodes$'
  RES_NS='^ +- namespaces$'
  RES_PODS='^ +- pods$'

  R="$(render)"
  for x in "$RES_NODES" "$RES_NS" "tokenreviews" "subjectaccessreviews" "operandrequests" "operatorgroups" "ibm-licensing-opreqs-role"; do
    assert_present "$x" "stock" "$R"
  done

  R="$(render --set ibmLicensing.spec.features.nodeCpuCappingEnabled=false)"
  assert_absent  "$RES_NODES"   "noNodeCap" "$R"
  assert_present "$RES_NS"      "noNodeCap" "$R"
  assert_present "tokenreviews" "noNodeCap" "$R"

  R="$(render --set ibmLicensing.spec.features.kubeRBACAuthEnabled=false)"
  assert_absent  "tokenreviews"          "noKubeAuth" "$R"
  assert_absent  "subjectaccessreviews"  "noKubeAuth" "$R"
  assert_present "$RES_NODES"            "noKubeAuth" "$R"

  R="$(render --set ibmLicensing.spec.features.operandRequestsEnabled=false)"
  assert_absent  "operandrequests"          "noOpreq" "$R"
  assert_absent  "operatorgroups"           "noOpreq" "$R"
  assert_absent  "ibm-licensing-opreqs-role" "noOpreq" "$R"
  assert_present "$RES_NODES"               "noOpreq" "$R"

  R="$(render --set ibmLicensing.spec.features.nssEnabled=true)"
  assert_present "$RES_NODES" "nssEnabled" "$R"
  # cluster namespaces list removed from operand roles, operator's own kept
  if command -v yq >/dev/null 2>&1; then
    ls_core="$(printf '%s' "$R" | yq 'select(.kind=="ClusterRole" and .metadata.name=="ibm-license-service") | .rules[] | select(.apiGroups[]=="") | .resources' 2>/dev/null)"
    if printf '%s' "$ls_core" | grep -q 'namespaces'; then
      bad "render[nssEnabled]: operand role still lists cluster namespaces"
    else
      ok "render[nssEnabled]: operand role drops cluster namespaces"
    fi
    op_ns="$(printf '%s' "$R" | yq 'select(.kind=="ClusterRole" and .metadata.name=="ibm-licensing-operator") | .rules[] | select(.resources[]=="namespaces") | .resources' 2>/dev/null)"
    if printf '%s' "$op_ns" | grep -q 'namespaces'; then
      ok "render[nssEnabled]: operator role keeps its namespaces[get]"
    else
      bad "render[nssEnabled]: operator role lost its namespaces[get]"
    fi
  fi

  # all off: restricted role's namespaces+nodes rule fully gone, no empty resources:
  R="$(render \
      --set ibmLicensing.spec.features.nodeCpuCappingEnabled=false \
      --set ibmLicensing.spec.features.kubeRBACAuthEnabled=false \
      --set ibmLicensing.spec.features.operandRequestsEnabled=false \
      --set ibmLicensing.spec.features.nssEnabled=true)"
  assert_present "$RES_PODS" "allOff" "$R"
  if command -v yq >/dev/null 2>&1; then
    if printf '%s' "$R" | yq 'true' >/dev/null 2>&1; then
      ok "render[allOff]: output is valid YAML (no empty resources:)"
    else
      bad "render[allOff]: output is invalid YAML"
    fi
  fi
fi

echo "------------------------------------------------------------"
echo "passed: ${pass}  failed: ${fail}"
[ "$fail" -eq 0 ]
