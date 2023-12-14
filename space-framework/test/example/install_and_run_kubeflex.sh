#!/usr/bin/env bash

# Create a kubeflex server.

git clone git@github.com:kubestellar/kubeflex.git
cd kubeflex
make
bin/kflex init
cd ..

