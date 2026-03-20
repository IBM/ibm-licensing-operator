<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Contributing guidelines](#contributing-guidelines)
    - [Developer Certificate of Origin](#developer-certificate-of-origin)
    - [Contributing A Patch](#contributing-a-patch)
    - [Issue and Pull Request Management](#issue-and-pull-request-management)
    - [Linting prerequisite](#linting-prerequisite)
    - [Pre-check before submitting a PR](#pre-check-before-submitting-a-pr)
    - [Build images](#build-images)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Contributing guidelines

## Developer Certificate of Origin

This repository built with [probot](https://github.com/probot/probot) that enforces
the [Developer Certificate of Origin](https://developercertificate.org/) (DCO) on Pull Requests. It requires all commit
messages to contain the `Signed-off-by` line with an email address that matches the commit author.

## Contributing A Patch

1. Submit an issue describing your proposed change to the repo in question.
1. The [repo owners](OWNERS) will respond to your issue promptly.
1. Fork the desired repo, develop and test your code changes.
1. Commit your changes with DCO
1. Submit a pull request.

## Issue and Pull Request Management

Anyone may comment on issues and submit reviews for pull requests. However, in
order to be assigned an issue or pull request, you must be a member of the
[IBM](https://github.com/ibm) GitHub organization.

Repo maintainers can assign you an issue or pull request by leaving a
`/assign <your Github ID>` comment on the issue or pull request.

## Linting prerequisite

- git
- go version v1.26+
- All linting tools are automatically installed when running `make install-linters`

| linting tool                                                            | version                                                                      | notes                                                                                                                             |
|-------------------------------------------------------------------------|------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------|
| [shellcheck](https://github.com/koalaman/shellcheck#installing)         | [v0.11.0](https://github.com/koalaman/shellcheck/releases/tag/v0.11.0)       | Automatically installed by `make install-linters`                                                                                 |
| [yamllint](https://github.com/adrienverge/yamllint#installation)        | [1.37.1](https://github.com/adrienverge/yamllint/releases/tag/v1.37.1)       | Automatically installed by `make install-linters` (requires Python with pip)                                                      |
| [golangci-lint](https://github.com/golangci/golangci-lint#install)      | [v2.11.2](https://github.com/golangci/golangci-lint/releases/tag/v2.11.2)    | Automatically installed by `make install-linters`                                                                                 |
| [mdl](https://github.com/markdownlint/markdownlint#installation)        | [0.15.0](https://github.com/markdownlint/markdownlint/releases/tag/v0.15.0)  | Automatically installed by `make install-linters` (requires Ruby with gem and bundler)                                            |
| [goimports](https://pkg.go.dev/golang.org/x/tools@v0.43.0/cmd/goimports) | [v0.43.0](https://pkg.go.dev/golang.org/x/tools@v0.43.0/cmd/goimports)       | Automatically installed by `make install-all-tools`                                                                               |

To install all required linters for the development process, run:

```shell
make install-linters
```

Some tools might need root privileges, so provide your password upon being asked.

- if you have an error during `make check`, for example:

```shell
goboringcrypto.h fatal error: openssl/ossl_typ.h: no such file or directory
```

Then try downloading newer golang version from [golang.org](https://golang.org) and:

- make sure $GOROOT will be set to the newer one

It was tested to work with these environment variables and setup:

- go version go1.26 linux/amd64
- Red Hat Enterprise Linux 8
- GOROOT=/usr/local/go
- GOPATH=$HOME/go
- GO111MODULE=on

## Development Workflow

### Quick development check

For a quick check during development, run:

```shell
make code-dev
```

This runs: `go mod tidy`, `go fmt`, `go vet`, and `make check` (all linters).

## Pre-check before submitting a PR

Before committing your changes, ensure you run the following checks:

### 1. Run linting checks

```shell
make check
```

This runs all linters (shellcheck, yamllint, golangci-lint, mdl) and go vet. This is also automatically run by git hooks before pushing code to remote branch.

### 2. Scan for secrets

```shell
make audit
```

This runs detect-secrets to ensure no secrets or sensitive credentials are present in the codebase. Always run this before committing to avoid accidentally exposing sensitive information.

### 3. After API changes

If you made changes to the API (files in `api/` or `controllers/` directories), you must regenerate the manifests and bundle:

```shell
make bundle
```

This command automatically runs:
- `make generate` - Generates deepcopy code for API types
- `make manifests` - Generates CRDs, RBAC, and webhook configurations
- Updates bundle manifests with the latest changes

**Important:** Always commit the generated files (CRDs, bundle manifests) along with your API changes.

## Building tools

Building the operators requires following tools installed:
- operator-sdk-v1.42.1
- opm-v1.64.0
- controller-gen-v0.20.1
- kustomize-v5.8.1
- yq-v4.52.4

Tools can installed using make target:

```shell
make install-all-tools
```

Toverify completness of the installed tools and their versions use the target:

```shell
make verify-installed-tools
```

Furthermore all important make targets are described in help section. To see the list of available commands run:

```shell
make help
```

## Build images

Make sure your code build passed.

```shell
make build
```

Now, you can follow the [README](./README.md) to work with the ibm-licensing-operator.

## Version bump

Run script `common/scripts/next_csv.sh` in project root directory with parameters: a current version, new version, old
version..
Example to bump operator from 4.2.20 to 4.2.21:

```shell
common/scripts/next_csv.sh 4.2.20 4.2.21 4.2.19
```

## Commit hook on OSX

When committing using your IDE doesn't work and commit hook fails, try extending PATH, for example I just opened
terminal from an IDE typed `echo $PATH` and pasted the result before running lint
in `common/scripts/.githooks/pre-commit` like so:

```shell
# in terminal where pre commit hooks are successful:
echo $PATH

# paste the result in pre-commit file, example:
...
PATH=$PATH:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/go/bin:...other paths...

.git/hooks/make_lint-all.sh
...
```

After that save this change to your local changelist in IDE, so you wont push it with Default (or other) changelist
changes.
