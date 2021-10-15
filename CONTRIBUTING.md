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

This repository built with [probot](https://github.com/probot/probot) that enforces the [Developer Certificate of Origin](https://developercertificate.org/) (DCO) on Pull Requests. It requires all commit messages to contain the `Signed-off-by` line with an email address that matches the commit author.

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

    | linting tool | version | instructions |
    | ------------ | ------- | ------------ |
    | [hadolint](https://github.com/hadolint/hadolint#install) | [v1.17.2](https://github.com/hadolint/hadolint/releases/tag/v1.17.2) | download binary from version link, make executable with `chmod +x` and add to bin directory |
    | [shellcheck](https://github.com/koalaman/shellcheck#installing) | [v0.7.0](https://github.com/koalaman/shellcheck/releases/tag/v0.7.0) | download binary from version link, make executable with `chmod +x` and add to bin directory |
    | [yamllint](https://github.com/adrienverge/yamllint#installation) | [v1.17.0](https://github.com/adrienverge/yamllint/releases/tag/v1.17.0) | download zip from version link, unzip and enter directory, then use python pip to install e.g. with this command `sudo pip3 install .` |
    | [golangci-lint](https://github.com/golangci/golangci-lint#install) | [v1.18.0](https://github.com/golangci/golangci-lint/releases/tag/v1.18.0) | `go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.18.0` |
    | [mdl](https://github.com/markdownlint/markdownlint#installation) | [v0.5.0](https://github.com/markdownlint/markdownlint/releases/tag/v0.5.0) | download using `git clone https://github.com/markdownlint/markdownlint.git -b v0.5.0` and install using `rake install` |
    | [awesome_bot](https://github.com/dkhamsing/awesome_bot#installation) | [1.19.1](https://github.com/dkhamsing/awesome_bot/releases/tag/1.19.1) | download using `git clone https://github.com/dkhamsing/awesome_bot.git -b 1.19.1` and install using `rake install` |
    | [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) | `3792095` | `go get golang.org/x/tools/cmd/goimports@3792095` |

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
make test
```

## Build images

Make sure your code build passed.

```shell
make
```

Now, you can follow the [README](./README.md) to work with the ibm-licensing-operator.

## Version bump

Run script `common/scripts/next_csv.sh` in project root directory with parameters: a current version, new version, old version..
Example to bump operator from 1.9.0 to 1.10.0:

```shell
common/scripts/next_csv.sh 1.9.0 1.10.0 1.8.0
```
