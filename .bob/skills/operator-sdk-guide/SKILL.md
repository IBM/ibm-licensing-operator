---
name: operator-sdk-guide
description: Explains how operator-sdk and kubebuilder work specifically for the ibm-licensing-operator - its CRDs, controllers, reconcile flow, scheme/manager wiring, kubebuilder markers, and the code-generation chain. Use to understand the operator's architecture, where to add a field or controller behavior, what a kubebuilder marker does here, or how the CRD/RBAC/bundle artifacts are generated.
---

# operator-sdk-guide

A repo-specific map of how this operator is built with the operator-sdk /
kubebuilder (`go.kubebuilder.io/v4`) framework. It is not a general SDK tutorial -
it describes *this* operator so you can navigate and extend it safely.

## Project shape

- Layout: `go.kubebuilder.io/v4`, domain `ibm.com`, group `operator` (see `PROJECT`).
- Module: `github.com/IBM/ibm-licensing-operator`, go 1.26.3.
- Entry point: `main.go` builds the runtime `scheme`, creates a controller-runtime
  `Manager`, and wires up the controllers.

## Custom Resources (owned CRDs)

| Kind | Package | Version | Scope | Types file |
|------|---------|---------|-------|-----------|
| IBMLicensing | `api/v1alpha1` | v1alpha1 | **cluster-scoped** | `ibmlicensing_types.go` |
| IBMLicensingMetadata | `api/v1alpha1` | v1alpha1 | namespaced | `ibmlicensingmetadata_types.go` |
| IBMLicensingDefinition | `api/v1` | v1 | namespaced | `ibmlicensingdefinition_types.go` |
| IBMLicensingQuerySource | `api/v1` | v1 | namespaced | `ibmlicensingquerysource_types.go` |

`IBMLicensing` is the primary CR: it defines a License Service instance (the operand).
`OperandRequest` (from ODLM) is watched but **external** - the operator reconciles it but
does not own its CRD (see `PROJECT` `external: true`).

Feature/config sub-types live in `api/v1alpha1/features/` (auth, alerting, hyper-threading,
prometheus query source) and helpers in `api/v1alpha1/{helper,license,features}.go`.

## Controllers (`controllers/`)

- **`ibmlicensing_controller.go`** - the main reconciler. `SetupWithManager` does
  `For(&IBMLicensing{})` and `Owns(...)` on the resources it manages: Deployment,
  Service, and (conditionally, when the APIs are present) Gateway API `Gateway`,
  `HTTPRoute`, `BackendTLSPolicy`. `Reconcile` drives many sub-reconcilers - secrets/tokens,
  ConfigMaps, Services, ServiceMonitors (RHMP + alerting), NetworkPolicy, Deployment,
  certificate secrets - and updates `.status`. It can also create a default instance
  (`CreateDefaultInstance`).
- **`operandrequest_controller.go`** + `operandrequest_discovery.go` - integrate with IBM
  ODLM OperandRequests.
- **`operatorgroup_cleaner.go`** - housekeeping for OperatorGroups.
- **`resources/`** - the builders that construct the operand's Kubernetes objects
  (containers, deployments, envs, services, CRDs, namespace scopes, operand bind info).
  This is where the actual desired-state objects are assembled.
- Tests: `*_controller_test.go` + `suite_test.go` (envtest/Ginkgo - see [[unit-test]]).

## Manager & scheme wiring (`main.go`)

`init()` registers every API group the operator touches into the runtime `scheme`:
client-go core, this operator's `v1alpha1`/`v1`, OpenShift `route` and `serviceca`,
Prometheus `monitoring`, `networking`, Red Hat Marketplace `meterdefinition`, ODLM, and
OperatorFramework. The manager honors `WATCH_NAMESPACE` (parsed via
`res.GetWatchNamespaceAsList()`) and `OPERATOR_NAMESPACE`. Each controller's
`SetupWithManager` is called to register it with the manager.

## kubebuilder markers (where behavior comes from)

Markers in Go comments drive code generation - edit the marker, then regenerate
(see [[generate-manifests]]):

- `// +kubebuilder:rbac:...` on the controller (e.g. lines around the `Reconcile` func in
  `ibmlicensing_controller.go`) generate `config/rbac/role.yaml`. Many here are
  `namespace=ibm-licensing` scoped plus a few cluster rules.
- `// +kubebuilder:object:root=true`, `subresource:status`, `printcolumn`, and validation
  markers in the `*_types.go` files drive the CRD schema and DeepCopy generation.
- `hack/boilerplate.go.txt` is the header injected into generated files.

## The generation chain

```
edit api/*_types.go or controller RBAC markers
        │
   make generate     ──▶ api/*/zz_generated.deepcopy.go   (controller-gen object)
   make manifests    ──▶ config/crd/bases/*, config/rbac/role.yaml, CSV base
   make bundle       ──▶ bundle/manifests/* (CSV+CRDs), bundle/metadata/*
```

`make bundle` additionally uses `operator-sdk generate kustomize manifests` + `kustomize`
+ `operator-sdk generate bundle`, then post-processes with `yq`: it forces IBMLicensing to
be the first owned CRD in the CSV, folds RBAC and `config/samples` CRs into the CSV
(`alm-examples`), and injects `common/relatedImages.yaml`. Always regenerate and commit
these together with API changes.

## Where to make common changes

| Goal | Edit | Then run |
|------|------|----------|
| Add a spec field | the relevant `api/**/…_types.go` | `make generate manifests` |
| Change validation / printer columns | kubebuilder markers in `_types.go` | `make manifests` |
| Grant the operator a new permission | `+kubebuilder:rbac` marker in the controller | `make manifests` |
| Change operand objects (Deployment/Service/…) | `controllers/resources/*.go` | `make unit-test` |
| Add reconcile behavior | `ibmlicensing_controller.go` (sub-reconcilers) | `make unit-test` |
| New CRD version bump for the CSV | version bump script + `make bundle` | see [[contributing]] |

## Related skills

- [[generate-manifests]] - run the generators after editing markers/types.
- [[unit-test]] - the envtest suites for the controllers described here.
- [[build-and-deploy]] - run the operator against a cluster.
- [[contributing]] - commit the generated artifacts with your change.
