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

FROM redhat/ubi9

ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM
ARG KUBESTELLAR_VERSION

WORKDIR /home/kubestellar

RUN groupadd kubestellar && \
    adduser -g kubestellar kubestellar && \
    mkdir -p .kcp kubestellar-logs && \
    dnf install -y jq procps && \
    curl -SL -o /usr/local/bin/yq "https://github.com/mikefarah/yq/releases/latest/download/yq_${TARGETOS}_${TARGETARCH}" && \
    chmod +x /usr/local/bin/yq && \
    curl -SL -o /usr/local/bin/kubectl "https://dl.k8s.io/release/v1.25.3/bin/${TARGETPLATFORM}/kubectl" && \
    chmod +x /usr/local/bin/kubectl && \
    curl -SL -o easy-rsa.tar.gz "https://github.com/OpenVPN/easy-rsa/releases/download/v3.1.5/EasyRSA-3.1.5.tgz" && \
    mkdir easy-rsa && \
    tar -C easy-rsa -zxf easy-rsa.tar.gz --wildcards --strip-components=1 EasyRSA*/* && \
    rm easy-rsa.tar.gz && \
    curl -SL -o kcp.tar.gz "https://github.com/kcp-dev/kcp/releases/download/v0.11.0/kcp_0.11.0_${TARGETOS}_${TARGETARCH}.tar.gz" && \
    mkdir kcp && \
    tar -C kcp -zxf kcp.tar.gz && \
    rm kcp.tar.gz && \
    curl -SL -o kcp-plugins.tar.gz "https://github.com/kcp-dev/kcp/releases/download/v0.11.0/kubectl-kcp-plugin_0.11.0_${TARGETOS}_${TARGETARCH}.tar.gz" && \
    mkdir kcp-plugins && \
    tar -C kcp-plugins -zxf kcp-plugins.tar.gz && \
    rm kcp-plugins.tar.gz && \
    curl -SL -o kubestellar.tar.gz "https://github.com/kubestellar/kubestellar/releases/download/${KUBESTELLAR_VERSION}/kubestellar_${KUBESTELLAR_VERSION}_${TARGETOS}_${TARGETARCH}.tar.gz" && \
    mkdir kubestellar && \
    tar -C kubestellar -zxf kubestellar.tar.gz && \
    rm kubestellar.tar.gz

ADD entry.sh            entry.sh
ADD kubestellar-patch   kubestellar/bin

RUN chown -R kubestellar:0 /home/kubestellar && \
    chmod -R g=u /home/kubestellar

# List volumes of ineterest
VOLUME [ "/home/kubestellar/.kcp" , "/home/kubestellar/kubestellar-logs" ]

# Setup the environment variables
ENV PATH=$PATH:/home/kubestellar/easy-rsa:/home/kubestellar/kcp/bin:/home/kubestellar/kcp-plugins/bin:/home/kubestellar/kubestellar/bin
ENV KUBECONFIG=/home/kubestellar/.kcp/admin.kubeconfig
ENV EXTERNAL_HOSTNAME=""
ENV EXTERNAL_PORT=""

# Switch the user
USER kubestellar

# start KubeStellar
CMD [ "/home/kubestellar/entry.sh" ]
