# Conditional RBAC configuration (cluster-scoped Helm chart)

The cluster-scoped IBM License Service Helm chart
(`deploy/argo-cd/components/license-service/helm-cluster-scoped`) can trim the
ClusterRole/Role RBAC it installs to match the optional operator features you
actually use. This is controlled by ILS-2352 (`_helpers.tpl` gates +
`conditionalize-helm-rbac.sh`).

**Defaults preserve behavior.** Every switch ships set to its *enabled/present*
default in `values.yaml`, reproducing previous releases' runtime behavior. RBAC
is only ever **trimmed** when you turn a feature off — nothing new is ever
granted.

> The defaults are written out explicitly (`features.nodeCpuCappingEnabled`/
> `kubeRBACAuthEnabled`/`operandRequestsEnabled: true`, `features.nssEnabled:
> false`) rather than left unset. The rendered **RBAC** is byte-identical to
> previous releases, and the operator treats explicit-default identically to
> absent (its `IsKubeRBACAuthEnabled` / `IsNodeCpuCappingEnabled` /
> `IsOperandRequestsEnabled` helpers default to `true`; `IsNamespaceScopeEnabled`
> defaults to `false` ⇒ cluster-wide discovery). The only manifest difference
> from earlier releases is that the rendered `IBMLicensing` CR now carries these
> fields in `spec` explicitly.

## Switches

Each switch is a single knob: the value both sets the field on the `IBMLicensing`
custom resource (via `cr.yaml`'s `toYaml (mergeOverwrite .Values.ibmLicensing.spec ...)`)
and flips the matching RBAC gate (via the `ibm-licensing.*` helpers in
`templates/_helpers.tpl`, which read the same `ibmLicensing.spec`). You never set
the CR field and the RBAC gate separately.

They live under `ibmLicensing.spec` in `values.yaml`, shipped at their
behavior-preserving defaults:

```yaml
ibmLicensing:
  spec:
    features:
      nssEnabled: false            # true scopes the operand to the watchNamespace set / nssConfigMap and disables cluster-wide namespace discovery
      nodeCpuCappingEnabled: true
      kubeRBACAuthEnabled: true
      operandRequestsEnabled: true
```

| values.yaml switch | CR spec field set | Default | RBAC removed when disabled / set |
|---|---|---|---|
| `ibmLicensing.spec.features.nodeCpuCappingEnabled: false` | `spec.features.nodeCpuCappingEnabled` | `true` (node access) | `nodes` resource dropped from the core-group rule of the `ibm-license-service` and `ibm-license-service-restricted` ClusterRoles |
| `ibmLicensing.spec.features.kubeRBACAuthEnabled: false` | `spec.features.kubeRBACAuthEnabled` | `true` (token/SAR review) | `authentication.k8s.io/tokenreviews` (create) and `authorization.k8s.io/subjectaccessreviews` (create) rules dropped from both operand ClusterRoles |
| `ibmLicensing.spec.features.operandRequestsEnabled: false` | `spec.features.operandRequestsEnabled` | `true` (ODLM integration) | `operator.ibm.com/operandrequests*` rule dropped from the `ibm-licensing-operator` ClusterRole; `operators.coreos.com/operatorgroups` rule dropped from the `ibm-licensing-operator` Role; the entire `cluster-rbac-for-operandrequests.yaml` content (`ibm-licensing-opreqs-role` ClusterRole + binding) is removed |
| `ibmLicensing.spec.features.nssEnabled: true` | `spec.features.nssEnabled` | `false` ⇒ cluster-wide discovery on | cluster-wide `namespaces` (get/list) dropped from the core-group rule of the `ibm-license-service` and `ibm-license-service-restricted` ClusterRoles. The operand is scoped to the `WATCH_NAMESPACE` set (or the ConfigMap named by `spec.features.nssConfigMap`). The `ibm-licensing-operator` ClusterRole's own `namespaces [get]` rule is **not** affected. |

Notes:

- On the `ibm-license-service-restricted` ClusterRole the `namespaces`+`nodes`
  rule is the only core-group rule, so when **both** `nodeCpuCappingEnabled` and
  namespace discovery are off the whole rule is removed (an outer guard prevents
  an invalid empty `resources:` list).
- The switches compose freely; any combination renders valid RBAC.

## Generation flow (maintainers)

The per-rule RBAC guards in `templates/cluster-rbac.yaml` and `templates/rbac.yaml`
are **generated**, not hand-edited. They are injected by
`common/scripts/conditionalize-helm-rbac.sh` (driven by the declarative
`common/makefile-generate/helm-rbac-gating-table`) at the end of
`make generate-yaml-argo-cd`, before the `createRBAC` wrap. Do **not** hand-edit
guards into these generated files — re-running the generator would overwrite
them, and CI would flag the drift.

`templates/cluster-rbac-for-operandrequests.yaml` has no kustomize counterpart;
its `operandRequestsEnabled` guard is maintained by the same script (whole-file
mode) and kept honest by the CI check.

CI / local verification:

- `make verify/helm-conditional-rbac` — fails if the committed templates' guards
  drift from what the script would produce (idempotent re-check).
- `make test/helm-conditional-rbac` — runs the fixture, idempotency and
  `helm template` render-matrix tests
  (`common/scripts/tests/conditionalize-helm-rbac-test.sh`).
