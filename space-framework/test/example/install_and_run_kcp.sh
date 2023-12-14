#!/usr/bin/env bash

# Create a kcp server.

git clone https://github.com/kcp-dev/kcp
cd kcp
make
bin/kcp start &> /tmp/kcp.log &
cd ..
