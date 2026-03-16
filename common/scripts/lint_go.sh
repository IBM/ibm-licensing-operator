#!/bin/bash
#
# Copyright 2023 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Populate $GOPATH/pkg/mod so golangci-lint's SSA builder does not have to
# re-extract all module sources.  The Makefile sets GOPATH=$(PWD)/.go which
# differs from the path used by plain `go build`, so the local module cache
# may be empty on first use after a fresh checkout or after adding new
# dependencies.  `go mod download` is a no-op when everything is cached.
go mod download

GOGC=25 golangci-lint run --print-resources-usage --verbose --concurrency 1 -c ./common/config/.golangci.yml
