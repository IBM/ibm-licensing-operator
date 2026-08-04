---
name: unit-test
description: Run the ibm-licensing-operator controller unit/integration tests (Ginkgo suites in ./controllers) against a live Kubernetes cluster via envtest with USE_EXISTING_CLUSTER=true. Use to validate controller or API changes before committing or opening a PR.
---

# unit-test

Runs the controller test suites (`suite_test.go`, `ibmlicensing_controller_test.go`,
`operandrequest_controller_test.go`, …). These are integration-style tests: they need a
reachable Kubernetes cluster (`USE_EXISTING_CLUSTER=true`) with the required CRDs and
RBAC applied. The `prepare-unit-test` target sets all of that up for you.

## When to use

- After changing anything in `controllers/` or `api/`.
- Before committing controller/API changes or opening a PR.
- To reproduce a CI unit-test failure locally.

## Commands

Prepare the cluster and run the full suite:

```bash
make prepare-unit-test && make unit-test
```

`make unit-test` already depends on `prepare-unit-test`, so this also works:

```bash
make unit-test
```

Under the hood the tests run:

```bash
go test -v ./controllers/... -coverprofile cover.out -timeout 30m
```

Coverage is written to `cover.out`.

## What prepare-unit-test does

1. Creates namespaces `ibm-licensing` (NAMESPACE) and `opreq-ns` (OPREQ_TEST_NAMESPACE).
2. Creates the `artifactory-token` pull secret (needs `ARTIFACTORY_USERNAME` /
   `ARTIFACTORY_TOKEN`).
3. Applies the `IBMLicensing` CRD and namespace-scoped RBAC (Role, ServiceAccount,
   RoleBinding) rendered from `config/rbac`.
4. Downloads and applies dependency CRDs from upstream: Red Hat Marketplace
   MeterDefinitions, Prometheus ServiceMonitors, ODLM OperandRequests, and Gateway API
   (GatewayClasses, Gateways, HTTPRoutes, BackendTLSPolicies).

## Prerequisites

- A running cluster (Kind, Minikube, or a real/OpenShift cluster) with `kubectl`
  pointed at it. `prepare-unit-test` mutates this cluster.
- `ARTIFACTORY_USERNAME` and `ARTIFACTORY_TOKEN` exported (for the pull secret).
- Internet access (dependency CRDs are fetched via `curl`).

## Key environment variables

| Var | Default | Purpose |
|-----|---------|---------|
| `NAMESPACE` | `ibm-licensing` | Operator/test namespace |
| `OPREQ_TEST_NAMESPACE` | `opreq-ns` | OperandRequest test namespace |
| `OCP` | *(unset)* | Set when running against OpenShift |
| `IBM_LICENSING_IMAGE` | derived | Operand image under test |

## Notes

- Tests can take up to 30 minutes (`-timeout 30m`).
- Because it uses an existing cluster, leftover resources may persist between runs;
  a fresh namespace or cluster avoids state bleed.
- This is the local equivalent of the Tekton PR test stage.

## Related skills

- [[code-quality]] - run before tests.
- [[build-and-deploy]] - run the operator against a cluster manually.
- [[operator-sdk-guide]] - how the controllers under test are wired.
