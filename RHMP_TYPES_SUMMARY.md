# RHMP Types — Summary and Compatibility Notes

## Why types are vendored in `pkg/rhmp/`

`github.com/redhat-marketplace/redhat-marketplace-operator/v2` was removed from `go.mod`
during the Phase 2+3 upgrade. The upstream module transitively imports k8s.io packages
that no longer exist in k8s.io v0.31+:

- `k8s.io/api/auditregistration/v1alpha1` (removed in k8s 1.22)
- `k8s.io/apimachinery/pkg/util/clock` (removed in k8s 1.25)
- `github.com/googleapis/gnostic/OpenAPIv2` (moved)

`go mod tidy` could not complete with the upstream module present. The only options were
to stay on k8s 1.25 forever, or vendor the handful of types actually used and drop the
module. The types were vendored into `pkg/rhmp/`.

## How MeterDefinition types are used

The `rhmp.MeterDefinition` type is **not** embedded in the IBMLicensing CRD schema.
`IBMLicensingSpec` only contains a simple `RHMPEnabled *bool` field.

Instead, the controller creates and manages `MeterDefinition` objects as **independent
Kubernetes resources** (their own CRD: `marketplace.redhat.com/v1beta1`), similar to
how it creates `Deployments` or `Services`. The reconciler (`reconcileMeterDefinition`
in `controllers/ibmlicensing_controller.go`) creates 4 MeterDefinition objects when
`RHMPEnabled` is true.

The MeterDefinition CRD must therefore be **pre-installed on the cluster** by the
Red Hat Marketplace Operator. It is not part of the IBM Licensing Operator's own CRD
bundle. In `prepare-unit-test` (Makefile), the CRD is installed manually because the
test cluster does not have RHMP running.

## Why the Makefile pins a specific commit

The Makefile fetches the MeterDefinition CRD YAML from a pinned commit:

```
https://raw.githubusercontent.com/redhat-marketplace/redhat-marketplace-operator/674d4e57186b/v2/config/crd/bases/marketplace.redhat.com_meterdefinitions.yaml
```

This commit (`674d4e57186b`, pseudo-version `v2.0.0-20230228135942`) is the exact
version the vendored Go types in `pkg/rhmp/` were derived from. The CRD schema
installed in the test cluster must match the vendored types, because:

- The operator constructs `MeterDefinition` objects using the vendored field JSON tags
- Kubernetes validates those objects against the installed CRD schema
- If the installed CRD schema is newer it may have required fields or renamed fields
  that the vendored types do not satisfy, causing `client.Create` / `client.List` to fail

The pinned commit and the vendored types are **intentionally the same snapshot** and
must be updated together.

## Current compatibility status

The vendored types (`pkg/rhmp/v1beta1/types.go`, `pkg/rhmp/common/types.go`) were
verified against the CRD YAML at commit `674d4e57186b` — all fields match:

| CRD field | Go struct | Status |
|-----------|-----------|--------|
| `group`, `kind`, `meters`, `resourceFilters`, `installedBy` | `MeterDefinitionSpec` | ✓ |
| `metricId`, `aggregation`, `query`, `workloadType` (required) | `MeterWorkload` json tags | ✓ |
| `dateLabelOverride`, `valueLabelOverride`, `groupBy`, `without`, `period`, `label`, `unit`, `name`, `description`, `metricType` | optional `MeterWorkload` fields | ✓ |
| `namespace`, `ownerCRD`, `label`, `annotation`, `workloadType` | `ResourceFilter` | ✓ |

## TODO: Update RHMP types

The vendored types and the Makefile CRD URL are both pinned to **February 2023**
(`v2.0.0-20230228135942`). The latest known upstream pseudo-version is
`v2.0.0-20260302222345-0e3562b432b3` — approximately 3 years of potential drift.

**Action required:**

1. Fetch the latest MeterDefinition CRD YAML from the current `main` branch of
   `redhat-marketplace/redhat-marketplace-operator`
2. Diff the `v1beta1` schema against the current vendored types in `pkg/rhmp/`
3. Update `pkg/rhmp/v1beta1/types.go` and `pkg/rhmp/common/types.go` to match
   any added, renamed, or removed fields
4. Update the commit SHA in the Makefile `prepare-unit-test` target to the
   corresponding new commit
5. Both changes must land in the same commit to keep the Go types and test CRD
   schema in sync

Fields most likely to have drifted: `MeterWorkload` required fields (`metricId`,
`aggregation`, `query`), `workloadType` enum values, and any new required fields
added to `MeterDefinitionSpec`.
