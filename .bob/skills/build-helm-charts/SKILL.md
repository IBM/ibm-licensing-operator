---
name: build-helm-charts
description: Build the development Helm charts for the License Service family (License Service, License Service Reporter, License Service Scanner) used for the "License Service without an IBM Cloud Pak" / no-operator deployment path. Use to package and test Helm-based deployments after changing the chart sources under deploy/argo-cd, helm-migration, or helm-no-operator.
---

# build-helm-charts

Builds the development Helm charts for the three License Service components. These charts
back the "License Service without an IBM Cloud Pak" (no-operator) and GitOps/ArgoCD
deployment scenarios. Chart sources live under `deploy/argo-cd/components/*`,
`helm-migration/`, and `helm-no-operator/`; this skill packages them with dev image tags
(branch-based) for testing.

## When to use

- After editing chart sources (`deploy/argo-cd/components/…`, `helm-migration/`,
  `helm-no-operator/`).
- To validate Helm packaging before releasing new chart versions.
- To produce dev `.tgz` charts wired to development image tags.

## Commands

Build all component charts:

```bash
make build/helm-develop-all
```

Or a single component:

```bash
make build/helm-develop-ls     # License Service (cluster-scoped) + migration chart
make build/helm-develop-lsr    # License Service Reporter (namespace + cluster scoped)
make build/helm-develop-lss    # License Service Scanner (namespace + cluster scoped)
```

Each delegates to `common/scripts/build-helm-develop.sh`, rewriting image tags from
`:$(CSV_VERSION)` to `:$(GIT_BRANCH)` and packaging via `helm`. Charts are optionally
pushed to Artifactory when `ARTIFACTORY_TOKEN` is set.

## What each target produces

- **ls** - `ibm-licensing-cluster-scoped` + `ibm-licensing-migration`
- **lsr** - `ibm-license-service-reporter` (+ postgresql, UI, oauth2-proxy, operator
  images) + `ibm-license-service-reporter-cluster-scoped`
- **lss** - `ibm-license-service-scanner` (scanner + scanner-operator) +
  `ibm-license-service-scanner-cluster-scoped`

## Prerequisites

- `helm` and `yq` installed (see [[setup-tools]]).
- `ARTIFACTORY_TOKEN` exported if pushing charts to the scratch Helm repo.

## Key environment variables

| Var | Purpose |
|-----|---------|
| `CSV_VERSION` | Base chart/image version (default `4.2.24`) |
| `GIT_BRANCH` | Branch used for dev image tags |
| `CHART_DESTINATION_*` | Artifactory destinations per component |

## Notes

- Migration charts use `MIGRATION_CHART_VERSION` rather than `CSV_VERSION`.
- Production chart publishing is done by the Tekton pipeline; this skill is for local dev
  iteration and validation.

## Related skills

- [[generate-manifests]] - `make generate-yaml-argo-cd` regenerates the ArgoCD component
  YAMLs these charts are built from.
- [[build-and-deploy]] - the operator/OLM deployment path (alternative to Helm).
