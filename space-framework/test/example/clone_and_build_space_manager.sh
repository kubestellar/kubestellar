#!/usr/bin/env bash

git clone git@github.com:kubestellar/kubestellar.git
cd kubestellar/space-framework
make codegen
make build
cd ../..
