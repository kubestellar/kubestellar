# Copyright 2023 The KubeStellar Authors.
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

###############################################################################
# Builder image
###############################################################################
FROM ubuntu AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM
ARG KUBESTELLAR_VERSION=v0.3.1

WORKDIR /

# install requirements
RUN apt-get update && \
    apt-get install -y curl wget

# download kubectl binaries
RUN wget -nv "https://dl.k8s.io/release/v1.25.3/bin/${TARGETPLATFORM}/kubectl" -O kubectl && \
    chmod 0755 kubectl

# download kcp binaries
RUN curl -SL -o kcp.tar.gz https://github.com/kcp-dev/kcp/releases/download/v0.11.0/kcp_0.11.0_${TARGETOS}_${TARGETARCH}.tar.gz && \
    mkdir kcp && \
    tar -C kcp -zxf kcp.tar.gz

# download kcp plugins
RUN curl -SL -o kcp-plugins.tar.gz https://github.com/kcp-dev/kcp/releases/download/v0.11.0/kubectl-kcp-plugin_0.11.0_${TARGETOS}_${TARGETARCH}.tar.gz && \
    mkdir kcp-plugins && \
    tar -C kcp-plugins -zxf kcp-plugins.tar.gz

# download KubeStellar binaries
RUN curl -SL -o kubestellar.tar.gz https://github.com/kubestellar/kubestellar/releases/download/${KUBESTELLAR_VERSION}/kubestellar_${KUBESTELLAR_VERSION}_${TARGETOS}_${TARGETARCH}.tar.gz && \
    mkdir kubestellar && \
    tar -C kubestellar -zxf kubestellar.tar.gz

# add entry script
ADD . /

###############################################################################
# KubeStellar image
###############################################################################
FROM ubuntu

# install requirements (for running inside the container)
RUN apt-get update && \
    apt-get install -y jq

WORKDIR /

# copy binaries from the builder image
COPY --from=builder /kubectl         /usr/local/bin/kubectl
COPY --from=builder /kcp/bin         /kcp
COPY --from=builder /kcp-plugins/bin /kcp-plugins
COPY --from=builder /kubestellar     /kubestellar
COPY --from=builder /entry.sh        /entry.sh

# setup the environment variables
ENV PATH=$PATH:/kcp:/kcp-plugins:/kubestellar/bin
ENV KUBECONFIG=/.kcp/admin.kubeconfig

# start KubeStellar
CMD [ "/entry.sh" ]