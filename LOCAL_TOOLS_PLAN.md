# Plan: Local Project-Directory Tool Installation

## Goal

Change all build tool installations so that:
- **Go** remains installed **system-wide** (no change)
- **All other tools** are installed into a **local project directory** (`./bin/`)
  instead of system-wide `/usr/local/bin/` or the GOPATH-based `.go/bin/`

This makes the project self-contained, avoids polluting the system PATH,
eliminates `sudo` requirements, and prevents version conflicts between projects.

---

## Current State Analysis

### Tool installation locations today

| Tool | Install method | Install location | Reference in Makefile |
|------|---------------|-----------------|----------------------|
| **Go** | System install | system-wide | `go` on PATH |
| **controller-gen** | `go install` | `$(GOBIN)` = `.go/bin/` | `$(shell which controller-gen)` or `$(GOBIN)/controller-gen` |
| **kustomize** | `go install` | `$(GOBIN)` = `.go/bin/` | `$(shell which kustomize)` or `$(GOBIN)/kustomize` |
| **opm** | `git clone` + `go build` | `~/opm` (home dir!) | `$(shell which opm)` or `~/opm` |
| **operator-sdk** | `curl` binary download | `/usr/local/bin/operator-sdk` | `operator-sdk` on PATH (via `command -v`) |
| **yq** | `curl` binary download | `/usr/local/bin/yq` | `yq` on PATH (via `command -v`) |
| **golangci-lint** | `curl` install script | `/usr/local/bin/golangci-lint` | `golangci-lint` on PATH |
| **hadolint** | `curl` binary / `brew` | `/usr/local/bin/hadolint` | `hadolint` on PATH |
| **shellcheck** | `curl` tarball / `brew` | `/usr/local/bin/shellcheck` | `shellcheck` on PATH |
| **goimports** | `go install` | `$(GOBIN)` = `.go/bin/` | `goimports` on PATH |
| **detect-secrets** | `pip install` | Python site-packages (system-wide) | `detect-secrets` on PATH |
| **yamllint** | `pip install` | Python site-packages | `yamllint` on PATH |
| **mdl** | `gem install` | Ruby gems | `mdl` on PATH |
| **awesome_bot** | `gem install` | Ruby gems | `awesome_bot` on PATH |

### Current problems

1. **Scattered locations** — tools end up in `/usr/local/bin/`, `~/`, `.go/bin/`,
   or language-specific package managers, making cleanup and version tracking difficult.
2. **`sudo` required** — install scripts use `/usr/local/bin/` which requires root on
   many systems. The `gem install` calls also use `sudo`.
3. **`which` fallback hides version mismatches** — the controller-gen and kustomize
   targets use `$(shell which ...)` first. If a system-wide version of a different
   version exists, it silently uses that instead of the project-required version.
4. **opm installs to home directory** — `cp ./bin/opm ~/` is unusual and not
   project-scoped.
5. **`catalogsource` targets download yq inline** — lines 479 and 498 download `yq`
   into `./yq` (project root), creating a temporary binary that isn't cleaned up and
   isn't the same as the `yq` used by other targets.
6. **GOPATH inside the project** — `GOPATH=$(PWD)/.go` already localizes Go module
   cache and `GOBIN`, but the `.go/bin/` directory mixes project tools with Go's
   own toolchain binaries.

---

## Target State

### Local tools directory: `./bin/`

All non-Go tools will be installed into `./bin/` within the project directory.
This directory is already in `.gitignore`.

Go tools installed via `go install` will also target `./bin/` by setting
`GOBIN=$(PWD)/bin`.

### Variable definitions (new)

```makefile
# Local bin directory for all project tools
LOCALBIN := $(PWD)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# Tool binaries (all in LOCALBIN)
CONTROLLER_GEN := $(LOCALBIN)/controller-gen
KUSTOMIZE      := $(LOCALBIN)/kustomize
OPM            := $(LOCALBIN)/opm
OPERATOR_SDK   := $(LOCALBIN)/operator-sdk
YQ             := $(LOCALBIN)/yq
GOLANGCI_LINT  := $(LOCALBIN)/golangci-lint
GOIMPORTS      := $(LOCALBIN)/goimports
DETECT_SECRETS := $(LOCALBIN)/detect-secrets

# Python venv for pip-installed tools (detect-secrets)
LOCALBIN_VENV  := $(LOCALBIN)/.venv
```

### PATH prepend

For targets that call tools by name (e.g., `operator-sdk generate kustomize manifests`
or linters), prepend `$(LOCALBIN)` to `PATH`:

```makefile
export PATH := $(LOCALBIN):$(PATH)
```

---

## Changes Required

### Phase A — Makefile restructure

#### A1. Define `LOCALBIN` and tool binary paths

Replace the current scattered tool-location logic with a single, deterministic
local bin directory.

**Remove:**
- The `GOPATH_DEFAULT` / `GOBIN_DEFAULT` / `GOPATH` / `GOBIN` / `DEST` block
  (lines 93–100). The Go module cache can stay in `.go/` but tool binaries
  should go to `./bin/`.
- The `$(GOBIN)` target and `work` target (lines 200–204).
- The `controller-gen:` conditional target (lines 554–567) — replace with
  deterministic download.
- The `kustomize:` conditional target (lines 569–582) — replace with
  deterministic download.
- The `opm:` conditional target (lines 584–600) — replace with prebuilt
  binary download.

**Add (after tool version definitions):**

```makefile
# Local bin directory for all project tools
LOCALBIN := $(PWD)/bin
export PATH := $(LOCALBIN):$(PATH)

# Keep GOPATH local for module cache, but install tool binaries into LOCALBIN
GOPATH_DEFAULT := $(PWD)/.go
export GOPATH ?= $(GOPATH_DEFAULT)
export GOBIN := $(LOCALBIN)
```

#### A2. Replace `controller-gen` target

**Before (lines 554–567):** conditional `which`-based resolution.

**After:**

```makefile
.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): $(LOCALBIN)
	@test -x $(CONTROLLER_GEN) && $(CONTROLLER_GEN) --version | grep -q "$(CONTROLLER_GEN_VERSION)" && echo "controller-gen $(CONTROLLER_GEN_VERSION) already installed" || \
		( echo "Installing controller-gen $(CONTROLLER_GEN_VERSION)..." && GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION) )
```

**How the "already installed" check works:**

All tool targets in this plan follow the same two-layer pattern:

1. **Make file-target check** — `$(CONTROLLER_GEN): $(LOCALBIN)` is a file
   target. Make will only evaluate the recipe if `./bin/controller-gen` does
   not exist on disk. If the binary is already present, Make skips the target
   entirely. This is the fast path for the common case.

2. **Version check in the recipe** — When the recipe does run (binary missing
   or explicitly requested via `make controller-gen`), it first checks whether
   the installed binary matches the required version. If it does, it prints
   "already installed" and exits. If the version doesn't match (or the binary
   doesn't exist), it re-installs.

This pattern is repeated for every tool (controller-gen, kustomize, opm,
operator-sdk, yq, golangci-lint, goimports).

#### A3. Replace `kustomize` target

**Before (lines 569–582):** conditional `which`-based resolution.

**After:**

```makefile
.PHONY: kustomize
kustomize: $(KUSTOMIZE)
$(KUSTOMIZE): $(LOCALBIN)
	@test -x $(KUSTOMIZE) && $(KUSTOMIZE) version | grep -q "$(KUSTOMIZE_VERSION)" && echo "kustomize $(KUSTOMIZE_VERSION) already installed" || \
		( echo "Installing kustomize $(KUSTOMIZE_VERSION)..." && GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION) )
```

#### A4. Replace `opm` target

**Before (lines 584–600):** `git clone` + `go build`, copies to `~/opm`.

**After:**

```makefile
.PHONY: opm
opm: $(OPM)
$(OPM): $(LOCALBIN)
	@test -x $(OPM) && $(OPM) version 2>/dev/null | grep -q "$(OPM_VERSION)" && echo "opm $(OPM_VERSION) already installed" || \
		( echo "Installing opm $(OPM_VERSION)..." && \
		  curl -sSfL https://github.com/operator-framework/operator-registry/releases/download/$(OPM_VERSION)/$(TARGET_OS)-$(LOCAL_ARCH)-opm -o $(OPM) && \
		  chmod +x $(OPM) )
```

This downloads the prebuilt binary directly into `./bin/opm`. No more
`git clone` + build from source. The version check prevents re-downloading
if the correct version is already present.

#### A5. Update `manifests`, `generate`, and all targets that use tool variables

Replace bare `$(CONTROLLER_GEN)` / `$(KUSTOMIZE)` / `$(OPM)` references
(already correct since they're variable references). The key change is that
these variables now point to `./bin/<tool>` instead of conditionally resolved
paths.

Targets that call tools by name on PATH (like `operator-sdk`, `yq`) will find
them via the `PATH := $(LOCALBIN):$(PATH)` prepend.

#### A6. Update `catalogsource` and `catalogsource-development` targets

**Before (lines 479, 498):**
```makefile
curl -Lo ./yq "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_$(TARGET_OS)_$(LOCAL_ARCH)"
chmod +x ./yq
./yq -i ...
```

**After:** Remove the inline `curl` download. Use `$(YQ)` (which is `./bin/yq`)
instead of `./yq`. Add `yq` as a dependency of the target.

```makefile
catalogsource: opm yq
	...
	$(YQ) -i ...
```

---

### Phase B — Install scripts update

#### B1. `common/scripts/install-operator-sdk.sh`

**Before:** Downloads to CWD, copies to `/usr/local/bin/operator-sdk`.

**After:** Downloads directly to the path provided as `$4` (the target binary path).

```bash
TARGET_OS=$1
LOCAL_ARCH=$2
OPERATOR_SDK_VERSION=$3
INSTALL_DIR=$4

echo ">>> Installing Operator SDK ${OPERATOR_SDK_VERSION} to ${INSTALL_DIR}"
mkdir -p "$(dirname "${INSTALL_DIR}")"
curl -sSfL "https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk_${TARGET_OS}_${LOCAL_ARCH}" \
  -o "${INSTALL_DIR}"
chmod +x "${INSTALL_DIR}"
```

Makefile target:

```makefile
.PHONY: operator-sdk
operator-sdk: $(OPERATOR_SDK)
$(OPERATOR_SDK): $(LOCALBIN)
	@test -x $(OPERATOR_SDK) && $(OPERATOR_SDK) version 2>/dev/null | grep -q "$(OPERATOR_SDK_VERSION)" && echo "operator-sdk $(OPERATOR_SDK_VERSION) already installed" || \
		( echo "Installing operator-sdk $(OPERATOR_SDK_VERSION)..." && \
		  bash common/scripts/install-operator-sdk.sh $(TARGET_OS) $(LOCAL_ARCH) $(OPERATOR_SDK_VERSION) $(OPERATOR_SDK) )
```

#### B2. `common/scripts/install-opm.sh`

**Before:** Downloads to CWD, copies to `/usr/local/bin/opm`.

**After:** Same pattern as operator-sdk — accept `$4` as target path.
Or eliminate the script entirely in favor of the inline `curl` in the
Makefile target (see A4). The script becomes redundant.

**Decision:** Remove `install-opm.sh` script; use inline `curl` in Makefile.

#### B3. `common/scripts/install-yq.sh`

**Before:** Downloads to CWD, copies to `/usr/local/bin/yq`.

**After:** Accept `$4` as target path.

```bash
TARGET_OS=$1
LOCAL_ARCH=$2
YQ_VERSION=$3
INSTALL_DIR=$4

echo ">>> Installing yq ${YQ_VERSION} to ${INSTALL_DIR}"
mkdir -p "$(dirname "${INSTALL_DIR}")"
curl -sSfL "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_${TARGET_OS}_${LOCAL_ARCH}" \
  -o "${INSTALL_DIR}"
chmod +x "${INSTALL_DIR}"
```

Makefile target:

```makefile
.PHONY: yq
yq: $(YQ)
$(YQ): $(LOCALBIN)
	@test -x $(YQ) && $(YQ) --version 2>/dev/null | grep -q "$(YQ_VERSION)" && echo "yq $(YQ_VERSION) already installed" || \
		( echo "Installing yq $(YQ_VERSION)..." && \
		  bash common/scripts/install-yq.sh $(TARGET_OS) $(LOCAL_ARCH) $(YQ_VERSION) $(YQ) )
```

#### B4. `common/scripts/install-linters-development.sh`

This script installs linters to `/usr/local/bin/` and uses `sudo gem install`.
It must be updated to install into `./bin/` instead.

Changes:
- Accept `LOCALBIN` as an environment variable or `$1` argument.
- Replace all `/usr/local/bin/` references with `${LOCALBIN}`.
- `golangci-lint`: use `-b "${LOCALBIN}"` flag in the install script.
- `hadolint`: download to `${LOCALBIN}/hadolint`.
- `shellcheck`: extract to `${LOCALBIN}/shellcheck`.
- `goimports`: `GOBIN="${LOCALBIN}" go install ...`.
- `mdl`, `awesome_bot`: These are Ruby gems and cannot easily be installed
  to a local `bin/`. Options:
  - Use `gem install --install-dir ./bin/.gems --bindir ./bin/` (if supported).
  - Or accept that Ruby linters remain system-wide (they are dev-only tools,
    not build-critical).
- `yamllint`: Python tool. Can use `pip install --target` or a venv.
  Options:
  - Use `python -m venv ./bin/.venv && ./bin/.venv/bin/pip install yamllint`
    and symlink `./bin/yamllint → .venv/bin/yamllint`.
  - Or accept that Python linters remain system-wide.

**Recommendation:** For `mdl`, `awesome_bot`, `yamllint`, and `detect-secrets`
(Python/Ruby tools), accept that they remain system-wide. They are development
linting tools, not build-critical. Focus local installation on the Go/binary
tools that directly affect build reproducibility.

#### B5. `common/scripts/install-detect-secrets.sh`

**Before:** Installs `detect-secrets` system-wide via `pip install` into
the global Python site-packages.

**After:** Creates a Python virtual environment in `$(LOCALBIN)/.venv/` and
installs `detect-secrets` there. A wrapper script at `$(LOCALBIN)/detect-secrets`
delegates to the venv binary, so it integrates seamlessly with the `PATH`
prepend.

Updated script:

```bash
#!/bin/bash
set -euo pipefail

LOCALBIN="${LOCALBIN:-.}"
VENV_DIR="${LOCALBIN}/.venv"

# Create venv if it doesn't exist
if [ ! -d "${VENV_DIR}" ]; then
  echo " » Creating Python virtual environment at ${VENV_DIR}"
  python3 -m venv "${VENV_DIR}"
fi

# Install or verify detect-secrets inside the venv
if "${VENV_DIR}/bin/detect-secrets" --version >/dev/null 2>&1; then
  echo " » detect-secrets already installed in venv"
  "${VENV_DIR}/bin/detect-secrets" --version
else
  echo " » Installing detect-secrets into ${VENV_DIR}"
  "${VENV_DIR}/bin/pip" install --upgrade \
    "git+https://github.com/ibm/detect-secrets.git@master#egg=detect-secrets"
fi

# Create wrapper script in LOCALBIN so detect-secrets is on PATH
cat > "${LOCALBIN}/detect-secrets" <<WRAPPER
#!/bin/bash
exec "${VENV_DIR}/bin/detect-secrets" "\$@"
WRAPPER
chmod +x "${LOCALBIN}/detect-secrets"
```

Makefile target:

```makefile
DETECT_SECRETS := $(LOCALBIN)/detect-secrets

.PHONY: install-detect-secrets
install-detect-secrets: $(DETECT_SECRETS)
$(DETECT_SECRETS): $(LOCALBIN)
	@test -x $(DETECT_SECRETS) && echo "detect-secrets already installed" || \
		LOCALBIN=$(LOCALBIN) bash common/scripts/install-detect-secrets.sh
```

The venv directory `./bin/.venv/` is covered by the existing `bin/` entry
in `.gitignore`.

---

### Phase C — Makefile `install-*` targets update

#### C1. `install-operator-sdk`

**Before:**
```makefile
install-operator-sdk:
	@operator-sdk version 2> /dev/null ; if [ $$? -ne 0 ]; then bash common/scripts/install-operator-sdk.sh ${TARGET_OS} ${LOCAL_ARCH} ${OPERATOR_SDK_VERSION}; fi
```

**After:**
```makefile
.PHONY: install-operator-sdk
install-operator-sdk: $(OPERATOR_SDK)
```

(Just an alias for the binary target.)

#### C2. `install-opm`

**Before:**
```makefile
install-opm:
	@opm version 2> /dev/null ; if [ $$? -ne 0 ]; then bash common/scripts/install-opm.sh ${TARGET_OS} ${LOCAL_ARCH} ${OPM_VERSION}; fi
```

**After:**
```makefile
.PHONY: install-opm
install-opm: $(OPM)
```

#### C3. `install-controller-gen`

**Before:**
```makefile
install-controller-gen:
	@controller-gen --version 2> /dev/null ; if [ $$? -ne 0 ]; then go install sigs.k8s.io/controller-tools/cmd/controller-gen@${CONTROLLER_GEN_VERSION}; fi
```

**After:**
```makefile
.PHONY: install-controller-gen
install-controller-gen: $(CONTROLLER_GEN)
```

#### C4. `install-kustomize`

**Before:**
```makefile
install-kustomize:
	@kustomize version 2> /dev/null ; if [ $$? -ne 0 ]; then go install sigs.k8s.io/kustomize/kustomize/v5@${KUSTOMIZE_VERSION}; fi
```

**After:**
```makefile
.PHONY: install-kustomize
install-kustomize: $(KUSTOMIZE)
```

#### C5. `install-yq`

**Before:**
```makefile
install-yq:
	@yq --version 2> /dev/null ; if [ $$? -ne 0 ]; then bash common/scripts/install-yq.sh ${TARGET_OS} ${LOCAL_ARCH} ${YQ_VERSION}; fi
```

**After:**
```makefile
.PHONY: install-yq
install-yq: $(YQ)
```

#### C6. `install-all-tools`

**Before:**
```makefile
install-all-tools: install-operator-sdk install-opm install-controller-gen install-kustomize install-yq verify-installed-tools install-detect-secrets
```

**After:** Same, but `verify-installed-tools` should be updated to check
`$(LOCALBIN)` instead of `command -v`.

#### C7. `install-linters`

**Before:**
```makefile
install-linters:
	common/scripts/install-linters-development.sh
```

**After:**
```makefile
install-linters:
	LOCALBIN=$(LOCALBIN) common/scripts/install-linters-development.sh
```

#### C8. `verify-installed-tools`

Update to check `$(LOCALBIN)/<tool>` existence instead of `command -v <tool>`.
Also print version from the local binary specifically:

```makefile
verify-installed-tools:
	@test -x $(OPERATOR_SDK) || { echo "Missing: operator-sdk. Run 'make install-all-tools'."; exit 1; }
	@test -x $(OPM) || { echo "Missing: opm. Run 'make install-all-tools'."; exit 1; }
	@test -x $(CONTROLLER_GEN) || { echo "Missing: controller-gen. Run 'make install-all-tools'."; exit 1; }
	@test -x $(KUSTOMIZE) || { echo "Missing: kustomize. Run 'make install-all-tools'."; exit 1; }
	@test -x $(YQ) || { echo "Missing: yq. Run 'make install-all-tools'."; exit 1; }
	@echo "All tools present in $(LOCALBIN)."
```

---

### Phase D — Update tool references throughout Makefile

#### D1. Replace bare `yq` calls with `$(YQ)`

The `manifests` target (lines 369–370) calls `yq` directly:
```makefile
yq -i '.metadata...' ...
```

Change to `$(YQ) -i ...` and add `yq` as a prerequisite.

Similarly, all targets that call `yq` by name:
- `manifests` (line 369, 370)
- `update-roles-alm-example` (lines 387–417)
- `alm-example` (lines 436–440)
- `generate-yaml-argo-cd` (lines 614–678)

For targets that call `operator-sdk` by name:
- `pre-bundle` (line 447, 448, 460)
- `scorecard` (line 475)

Since `export PATH := $(LOCALBIN):$(PATH)` is set at the top of the Makefile,
bare `yq` and `operator-sdk` calls will resolve to `./bin/yq` and
`./bin/operator-sdk`. However, **best practice** is to use `$(YQ)` and
`$(OPERATOR_SDK)` variables for explicitness and to enable Make dependency
tracking.

#### D2. Replace bare `detect-secrets` calls with `$(DETECT_SECRETS)`

The `audit` target (line 235–237) calls `detect-secrets` directly:
```makefile
audit: install-detect-secrets
	@detect-secrets scan ...
	@detect-secrets audit ...
```

Change to `$(DETECT_SECRETS) scan ...` / `$(DETECT_SECRETS) audit ...`.
The `install-detect-secrets` prerequisite already ensures the venv and
wrapper are created.

#### D3. Remove the old `work` / `$(GOBIN)` targets

These become unnecessary since `LOCALBIN` replaces their purpose.

---

### Phase E — `.gitignore` verification

The `.gitignore` already contains `bin/` (line 81), so all locally installed
tools will be excluded from version control, including the Python venv at
`./bin/.venv/`. No changes needed.

---

### Phase F — Linter tools (secondary priority)

For `golangci-lint`, `hadolint`, `shellcheck`, and `goimports`:

```makefile
GOLANGCI_LINT := $(LOCALBIN)/golangci-lint
HADOLINT      := $(LOCALBIN)/hadolint
SHELLCHECK    := $(LOCALBIN)/shellcheck
GOIMPORTS     := $(LOCALBIN)/goimports

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(LOCALBIN)
	@test -x $(GOLANGCI_LINT) && $(GOLANGCI_LINT) --version 2>/dev/null | grep -q "$(GOLANGCI_LINT_VERSION)" && echo "golangci-lint $(GOLANGCI_LINT_VERSION) already installed" || \
		( echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..." && \
		  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCALBIN) $(GOLANGCI_LINT_VERSION) )

.PHONY: goimports
goimports: $(GOIMPORTS)
$(GOIMPORTS): $(LOCALBIN)
	@test -x $(GOIMPORTS) && echo "goimports already installed" || \
		( echo "Installing goimports $(GOIMPORTS_VERSION)..." && \
		  GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION) )
```

For `hadolint` and `shellcheck`, update `install-linters-development.sh` to
accept `LOCALBIN` and install there.

For `yamllint`, `mdl`, `awesome_bot` (Python/Ruby linting tools):
keep system-wide. These tools don't affect build artifact reproducibility.

---

## Summary of files to change

| File | Changes |
|------|---------|
| `Makefile` | Restructure `LOCALBIN`/`GOBIN`; replace all tool targets; update tool references; update `install-*` targets; remove old `which`-based logic |
| `common/scripts/install-operator-sdk.sh` | Accept target path as `$4`; remove `/usr/local/bin/` |
| `common/scripts/install-yq.sh` | Accept target path as `$4`; remove `/usr/local/bin/` |
| `common/scripts/install-opm.sh` | Remove (replaced by inline curl in Makefile) or update like operator-sdk |
| `common/scripts/install-detect-secrets.sh` | Rewrite to create venv at `$(LOCALBIN)/.venv/`, install detect-secrets there, create wrapper script at `$(LOCALBIN)/detect-secrets` |
| `common/scripts/install-linters-development.sh` | Accept `LOCALBIN` env var; install binary tools to `$LOCALBIN` instead of `/usr/local/bin/`; remove `sudo` from gem installs |

## Out of scope

| Tool | Reason |
|------|--------|
| **Go** | Explicitly required to stay system-wide |
| **yamllint** | Python pip tool — not build-critical, dev linting only |
| **mdl** | Ruby gem — not build-critical, dev linting only |
| **awesome_bot** | Ruby gem — not build-critical, dev linting only |
| **diffutils** | System package (`apt-get`/`brew`) — cannot be localized to `./bin/` |
| **docker/podman** | Container runtime — always system-wide |
| **kubectl** | Cluster CLI — always system-wide |

---

## Execution order

1. **Phase A** — Makefile restructure (core tool targets)
2. **Phase B** — Update install scripts
3. **Phase C** — Update `install-*` convenience targets
4. **Phase D** — Update bare tool name references to use variables
5. **Phase E** — Verify `.gitignore` (no changes expected)
6. **Phase F** — Linter tools (can be deferred if needed)

**Verification after each phase:**
- `make controller-gen` installs to `./bin/controller-gen`
- `make kustomize` installs to `./bin/kustomize`
- `make opm` installs to `./bin/opm`
- `make install-operator-sdk` installs to `./bin/operator-sdk`
- `make install-yq` installs to `./bin/yq`
- `make install-detect-secrets` creates venv at `./bin/.venv/` and wrapper at `./bin/detect-secrets`
- Running any `make <tool>` twice in a row skips re-installation ("already installed")
- `make generate manifests bundle` succeeds using local tools
- `make audit` uses local `./bin/detect-secrets` (not system-wide)
- No files written to `/usr/local/bin/`, `~/`, or `.go/bin/`
