---
name: build-and-deploy
description: Build the ibm-licensing-operator binary/image and run or deploy it against a Kubernetes cluster for local development and testing. Use to compile the operator, run the controller locally against a cluster, build and push a dev image to the scratch registry, or install CRDs and deploy the operator into a cluster.
---

# build-and-deploy

Local build-and-run loop for the operator. Covers three levels: compile the binary,
run the controller directly against a cluster (fastest iteration), and deploy it into a
cluster via kustomize. Production/multi-arch/catalog builds are handled by the Tekton
pipeline, not here.

## When to use

- Compile-check the operator after code changes.
- Run the controller locally against your kubeconfig cluster for fast iteration.
- Build and push a development image to the scratch registry to test in-cluster.
- Install the CRDs and deploy the operator into a test cluster.

## Commands

### Compile the binary

```bash
make build          # builds bin/ibm-licensing-operator for LOCAL_ARCH via gobuild.sh
```

### Run the controller locally (fastest iteration)

```bash
NAMESPACE=ibm-licensing make run
```

Runs `go run ./main.go` with `WATCH_NAMESPACE` / `OPERATOR_NAMESPACE` set to `NAMESPACE`
(default `ibm-licensing`). The controller runs on your machine but reconciles against the
cluster in `~/.kube/config`. CRDs must already be installed (`make install`).

### Build + push a development image

```bash
make build-push-image-development
```

Builds `docker build -f Dockerfile` for `LOCAL_ARCH`, tags it into the **scratch**
registry as `<IMAGE_NAME>-<arch>:<VERSION>` (and, on amd64, `:<GIT_BRANCH_TAG>`), pushes
it, and records the reference in `.published-images.txt`. The push is a plain
`docker push`, so your Docker/Podman client must be **logged in** to the scratch registry
first (`docker login <SCRATCH_REGISTRY host>` with your Artifactory username + API
token/key) - the target does not read any Artifactory env vars.

### Install CRDs and deploy into a cluster

```bash
make install        # kustomize build config/crd | kubectl apply -f -
make deploy         # sets manager image to $IMG, applies config/default
```

Teardown:

```bash
make uninstall      # removes the CRDs
```

## Key environment variables

| Var | Default | Purpose |
|-----|---------|---------|
| `NAMESPACE` | `ibm-licensing` | Watch/operator namespace for `run`/`deploy` |
| `IMG` | `ibm-licensing-operator` | Image reference used by `make deploy` |
| `LOCAL_ARCH` | host arch | Build target (amd64/ppc64le/s390x/arm64) |
| `SCRATCH_REGISTRY` | Makefile default | Dev registry for pushed images |
| `VERSION` | derived from `CSV_VERSION` | Image tag |

## Choosing an approach

- **Iterating on controller logic** → `make run` (no image build, instant feedback).
- **Testing the packaged image / RBAC / deployment** → `make build-push-image-development`
  then `make install && make deploy`.
- **Just checking it compiles** → `make build`.

## Prerequisites

- `kubectl` pointed at a reachable cluster; kustomize installed (see [[setup-tools]]).
- For image builds: Docker or Podman running, and (to push) the client logged in to the
  scratch registry via `docker login` using your Artifactory username + API token/key.
  The push targets use `docker push`, not env vars - unlike [[unit-test]], which reads the
  `ARTIFACTORY_USERNAME` / `ARTIFACTORY_TOKEN` pair to build a pull secret.
- Install the CRDs (`make install`) before `make run`.

## Notes

- Multi-arch, `latest`, and OLM catalog builds are release/pipeline concerns handled by
  Tekton - don't run them for local development.
- `make deploy` mutates `config/manager` kustomization to set the image; avoid committing
  that incidental change.

## Related skills

- [[generate-manifests]] - regenerate CRDs before installing them.
- [[unit-test]] - automated controller tests against a cluster.
- [[build-helm-charts]] - the Helm-based (no-operator) deployment path.
