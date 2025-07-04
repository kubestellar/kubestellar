#!/bin/bash

# Copyright 2024 The KubeStellar Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This file contains code derived from Envoy Gateway, https://github.com/envoyproxy/gateway
# and is provided here subject to the following: Copyright Envoy Gateway Authors SPDX-License-Identifier: Apache-2.0

# Required by `compile_native_go_fuzzer`
# Ref: https://google.github.io/oss-fuzz/getting-started/new-project-guide/go-lang/#buildsh
cd "$SRC"
git clone --depth=1 https://github.com/AdamKorcz/go-118-fuzz-build --branch=include-all-test-files
cd go-118-fuzz-build
go build .
mv go-118-fuzz-build /root/go/bin/

cd "$SRC"/kubestellar

set -o nounset
set -o pipefail
set -o errexit
set -x

# Create empty file that imports "github.com/AdamKorcz/go-118-fuzz-build/testing"
# This is a small hack to install this dependency, since it is not used anywhere,
# and Go would therefore remove it from go.mod once we run "go mod tidy && go mod vendor".
printf "package kubestellar\nimport _ \"github.com/AdamKorcz/go-118-fuzz-build/testing\"\n" > register.go
go mod edit -replace github.com/AdamKorcz/go-118-fuzz-build="$SRC"/go-118-fuzz-build
go mod tidy

# compile native-format fuzzers
compile_native_go_fuzzer github.com/kubestellar/kubestellar/test/fuzz FuzzJSONPathParsing FuzzJSONPathParsing
compile_native_go_fuzzer github.com/kubestellar/kubestellar/test/fuzz FuzzCRDValidation FuzzCRDValidation
compile_native_go_fuzzer github.com/kubestellar/kubestellar/test/fuzz FuzzLabelParsing FuzzLabelParsing
compile_native_go_fuzzer github.com/kubestellar/kubestellar/test/fuzz FuzzAPIGroupParsing FuzzAPIGroupParsing
compile_native_go_fuzzer github.com/kubestellar/kubestellar/test/fuzz FuzzJSONValueValidation FuzzJSONValueValidation

# add seed corpus
zip -j $OUT/FuzzJSONPathParsing_seed_corpus.zip "$SRC"/kubestellar/test/fuzz/testdata/*.yaml
zip -j $OUT/FuzzCRDValidation_seed_corpus.zip "$SRC"/kubestellar/test/fuzz/testdata/*.yaml
zip -j $OUT/FuzzLabelParsing_seed_corpus.zip "$SRC"/kubestellar/test/fuzz/testdata/*.yaml
zip -j $OUT/FuzzAPIGroupParsing_seed_corpus.zip "$SRC"/kubestellar/test/fuzz/testdata/*.yaml
zip -j $OUT/FuzzJSONValueValidation_seed_corpus.zip "$SRC"/kubestellar/test/fuzz/testdata/*.yaml 