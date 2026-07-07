---
name: generate-manifests
description: Regenerate all derived artifacts after changing the ibm-licensing-operator API types or config - DeepCopy code, CRDs, RBAC, the OLM bundle/CSV, and the ArgoCD GitOps YAMLs. Use after editing anything in api/ or controllers/ (new fields, kubebuilder markers, RBAC markers, samples) so generated files stay in sync and can be committed.
---

# generate-manifests

Whenever you touch the API (`api/v1`, `api/v1alpha1`) or RBAC/kubebuilder markers in
`controllers/`, several files are generated from them and **must be regenerated and
committed** together with the source change. This skill covers the whole generation
chain, from cheap codegen to the full OLM bundle.

## When to use

- Added/changed a field, type, or validation in `api/v1*/…_types.go`.
- Changed a `// +kubebuilder:...` marker (CRD schema, printer columns, RBAC).
- Edited a sample CR in `config/samples/` (feeds the CSV `alm-examples`).
- Need refreshed CRDs, RBAC role, CSV/bundle, or ArgoCD manifests before a PR.

## Commands (cheapest first)

### 1. Code + CRD/RBAC generation

```bash
make generate      # DeepCopy methods -> api/*/zz_generated.deepcopy.go
make manifests     # CRDs -> config/crd/bases, RBAC role, CSV base annotations
```

Or both at once:

```bash
make manifests generate
```

### 2. Full OLM bundle (do this before a PR that changes the API)

```bash
make bundle
```

`make bundle` runs `generate` + `manifests`, then via `pre-bundle`:
- `operator-sdk generate kustomize manifests` + `kustomize build config/manifests`
- `operator-sdk generate bundle` (writes `bundle/manifests`, `bundle/metadata`,
  `bundle/tests`)
- reorders owned CRDs so **IBMLicensing is first**, sets package/OpenShift-version
  annotations, injects `relatedImages` from `common/relatedImages.yaml`
- folds RBAC rules and `alm-examples` from `config/samples` into the CSV
- validates the bundle with `operator-sdk bundle validate ./bundle`

### 3. ArgoCD / GitOps YAMLs (only when the deployment shape changed)

```bash
make generate-yaml-argo-cd
```

Rebuilds the templated ArgoCD manifests (cluster-rbac, cr, crd, deployment, rbac,
serviceaccounts) under the `deploy/argo-cd` component sources from the kustomize output.
Needed only when the operator Deployment, RBAC, or CRDs change - not for pure API field
tweaks that don't alter RBAC.

## Files produced / updated

- `api/v1/zz_generated.deepcopy.go`, `api/v1alpha1/zz_generated.deepcopy.go`
- `config/crd/bases/operator.ibm.com_*.yaml`
- `config/rbac/role.yaml`
- `config/manifests/bases/ibm-licensing-operator.clusterserviceversion.yaml`
- `bundle/manifests/…` (CSV + CRDs), `bundle/metadata/annotations.yaml`

## The owned CRDs (what the CRDs cover)

| Kind | API version | Scope |
|------|-------------|-------|
| IBMLicensing | v1alpha1 | cluster-scoped |
| IBMLicensingMetadata | v1alpha1 | namespaced |
| IBMLicensingDefinition | v1 | namespaced |
| IBMLicensingQuerySource | v1 | namespaced |

## Prerequisites

controller-gen, kustomize, yq, and operator-sdk installed - see [[setup-tools]].

## Notes

- **Always commit generated files** with the API change (CONTRIBUTING.md requires it).
- `CSV_VERSION` (default `4.2.24`) drives the bundle version; overriding it changes CSV
  contents - don't bump it as a side effect of regeneration.
- The `temp/` directory holds intermediate build files and is gitignored; don't commit it.
- Version bumps use `common/scripts/next_csv.sh <current> <new> <old>`, not this skill.

## Related skills

- [[operator-sdk-guide]] - what the markers and generators actually do.
- [[code-quality]] / [[unit-test]] - run after regenerating.
- [[build-and-deploy]] - deploy the regenerated CRDs to a cluster.
