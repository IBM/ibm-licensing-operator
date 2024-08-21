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
- go version v1.17+
- some tools below requires python with pip (tested on python3), and ruby with gem and bundler to install
- Linting Tools

| linting tool                                                            | version                                                                      | instructions                                                                                                                             |
|-------------------------------------------------------------------------|------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------|
| [hadolint](https://github.com/hadolint/hadolint#install)                | [v2.12.0](https://github.com/hadolint/hadolint/releases/tag/v2.12.0)         | - download binary from version link, make executable with `chmod +x` and add to bin directory<br/>- for MacOS: `brew install hadolint`   |
| [shellcheck](https://github.com/koalaman/shellcheck#installing)         | [v0.8.0](https://github.com/koalaman/shellcheck/releases/tag/v0.8.0)         | - download binary from version link, make executable with `chmod +x` and add to bin directory<br/>- for MacOS: `brew install shellcheck` |
| [yamllint](https://github.com/adrienverge/yamllint#installation)        | [v1.28.0](https://github.com/adrienverge/yamllint/releases/tag/v1.28.0)      | - `pip install yamllint==1.28.0`                                                                                                         |
| [golangci-lint](https://github.com/golangci/golangci-lint#install)      | [v1.56.1](https://github.com/golangci/golangci-lint/releases/tag/v1.56.1)    | - `go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.1`<br/>- for MacOS: `brew install golangci-lint`                     |
| [mdl](https://github.com/markdownlint/markdownlint#installation)        | [v0.11.0](https://github.com/markdownlint/markdownlint/releases/tag/v0.11.0) | - download using `git clone https://github.com/markdownlint/markdownlint.git -b v0.11.0` and install using `sudo rake install`           |
| [awesome_bot](https://github.com/dkhamsing/awesome_bot#installation)    | [1.20.0](https://github.com/dkhamsing/awesome_bot/releases/tag/1.20.0)       | - download using `git clone https://github.com/dkhamsing/awesome_bot.git -b 1.20.0` and install using `sudo rake install`                |
| [goimports](https://pkg.go.dev/golang.org/x/tools@v0.3.0/cmd/goimports) | [v0.3.0](https://pkg.go.dev/golang.org/x/tools@v0.3.0/cmd/goimports)         | - `go install golang.org/x/tools/cmd/goimports@v0.3.0`                                                                                   |
| [diffutils](https://www.gnu.org/software/diffutils/)                    | [v3.8](https://ftp.gnu.org/gnu/diffutils/diffutils-3.8.tar.xz)               | - download binary from version link, make executable with `chmod +x` and add to bin directory<br/>- for MacOS: `brew install diffutils`  |

To install required linters for the development process, you can use script:

```shell
make install-linters
```

Some tools will need root privileges, so provide your password upon being asked.

- if you have an error during `make check`, for example:

```shell
goboringcrypto.h fatal error: openssl/ossl_typ.h: no such file or directory
```

Then try downloading newer golang version from [golang.org](https://golang.org) and:

- make sure $GOROOT will be set to the newer one

It was tested to work with these environment variables and setup:

- go version go1.17 linux/amd64
- Red Hat Enterprise Linux 8
- GOROOT=/usr/local/go
- GOPATH=$HOME/go
- GO111MODULE=on

## Pre-check before submitting a PR

After your PR is ready to commit, please run following commands to check your code.

```shell
make check
```

## Building tools

Building the operators requires following tools installed:
- operator-sdk-v1.25.2
- opm-v1.26.2
- controller-gen-v0.7.0
- kustomize-v4.5.7
- yq-v4.30.5

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
Example to bump operator from 1.9.0 to 1.10.0:

```shell
common/scripts/next_csv.sh 1.9.0 1.10.0 1.8.0
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