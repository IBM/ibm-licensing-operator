---
name: license-service-architecture
description: How License Service is split across three separate repositories - the operator (this repo), the operand (the containerized License Service application), and the commons (a shared library) - and how they interplay at build time and runtime. Use to understand which repo owns what, where a change belongs, and how the pieces fit together. Does NOT assume the repos are checked out together or live in the same directory/catalog.
---

# license-service-architecture

License Service is built from **three separate repositories**. This repo,
`ibm-licensing-operator`, is only one of them. Knowing the split tells you **where a change
belongs** and why some behavior can't be changed from here.

> These are independent repositories with their own release cycles. **Do not assume they are
> checked out side by side, in the same parent directory/catalog, or available locally at
> all.** They connect through published artifacts (a container image, a library jar) and
> through the CR/API contract - not through a shared source tree.

## The three repositories

### 1. Operator — `ibm-licensing-operator` (this repo)
- **Language/stack:** Go, built with operator-sdk / kubebuilder.
- **Role:** installs, configures, and manages License Service on a cluster. It owns the
  `IBMLicensing` CRD and the reconcile logic that turns a CR into running Kubernetes objects
  (Deployment, Services, Route/Ingress, ConfigMaps, certificates, RBAC, monitoring).
- **What it does NOT contain:** the License Service application logic itself. The operator
  deploys the operand's **image**; it does not build or import the operand's source.

### 2. Operand — the containerized License Service application
- **Language/stack:** Java / Spring Boot, multi-module Maven, shipped as a multi-arch
  container image.
- **Role:** the actual workload that does the licensing work - collecting, calculating, and
  serving/reporting license-usage data. This is what the operator's Deployment runs.
- **Relationship:** the operator references it **only as a container image** (via the
  licensing image env/config, e.g. `IBM_LICENSING_IMAGE`), pinned per release. Changing what
  License Service *does at runtime* is an operand change, not an operator change.

### 3. Commons — the shared licensing library
- **Language/stack:** a Java library, published as a jar (no container image of its own).
- **Role:** shared domain code (models/POJOs, persistence, shared API types and utilities)
  used by the operand. It is a **build-time dependency of the operand**, consumed as a
  published artifact.
- **Relationship:** the operator (Go) does **not** depend on commons at all.

## How they interplay

```
   licensing-commons (Java lib, jar)
            │  build-time dependency (published jar)
            ▼
   operand (Java app)  ──build──▶  License Service container image
            ▲                                   │
            │ deploys / configures the image     │ runtime
   ibm-licensing-operator (Go)  ───────────────▶ running License Service Deployment
   (reconciles the IBMLicensing CR)              (+ Services, Route, ConfigMaps, certs)
```

- **Commons → operand** is a **build-time** link: the operand pulls the commons jar as a
  dependency and compiles against it. Nothing in this direction touches the operator.
- **Operator → operand** is a **runtime/deployment** link: the operator never imports operand
  code; it deploys the operand's pre-built image and configures it through the CR spec and
  generated ConfigMaps/env. The contract between them is the **image + its configuration
  surface**, not shared source.
- **Operator ↔ commons:** no relationship. Different language, different toolchain.

The result is loose coupling: the operator and operand are versioned and released
independently, and the operand image is pinned into the operator's manifests per release.

## Where does my change belong?

- **Deploy/config/lifecycle** (CRD fields, how the Deployment/Service/Route/RBAC/monitoring
  are shaped, install flows, Helm/OLM packaging) → **operator** (this repo).
- **License Service behavior** (data collection, calculation, the APIs/endpoints it serves,
  reporting, alerting logic) → **operand** repo.
- **Shared domain models / persistence / shared types** used by the application →
  **commons** repo (then consumed by the operand).

A single user-visible feature can span repos: e.g. exposing a new operand setting usually
means an operand change (implement the behavior) **and** an operator change (surface it on the
`IBMLicensing` CR and wire it into the generated config). Plan such work as coordinated
changes across the relevant repos, released in the right order (commons → operand → operator).

## Working across repos

- Confirm the **operand image version** the operator pins for the release you're on before
  assuming operand behavior; operator and operand advance independently.
- If the other repos aren't available locally, don't fabricate their internals - reason from
  the contract (the CR spec, the image, the published jar) and, when in doubt, say a change
  needs to be made in the operand/commons repo rather than guessing its code.

## Related skills

- [[operator-sdk-guide]] - the internal architecture of **this** repo (the operator).
- [[generate-manifests]] - regenerating CRDs/bundle when the operator's CR/config surface
  changes to expose operand configuration.
- [[build-and-deploy]] - how the operator deploys the operand image locally.
