###############################################################################
# Builder image
###############################################################################
FROM --platform=$BUILDPLATFORM redhat/ubi9 AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM
ARG GIT_DIRTY=dirty

RUN groupadd kubestellar && useradd -g kubestellar kubestellar

WORKDIR /home/kubestellar

RUN mkdir -p .kcp && \
    dnf install -y git golang jq procps && \
    go install github.com/mikefarah/yq/v4@v4.34.2 && \
    curl -SL -o /usr/local/bin/kubectl "https://dl.k8s.io/release/v1.25.3/bin/${TARGETPLATFORM}/kubectl" && \
    chmod +x /usr/local/bin/kubectl && \
    curl -SL -o easy-rsa.tar.gz "https://github.com/OpenVPN/easy-rsa/releases/download/v3.1.5/EasyRSA-3.1.5.tgz" && \
    got_hash=$(sha256sum easy-rsa.tar.gz  | awk '{ print $1 }') && \
    if [ "$got_hash" != 9fc6081d4927e68e9baef350e6b3010c7fb4f4a5c3e645ddac901081eb6adbb2 ]; then \
       echo "Got bad copy of EasyRSA-3.1.5.tgz" >&2 ; \
       exit 1; \
    fi && \
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
    git config --global --add safe.directory /home/kubestellar && \
    mkdir -p bin && \
    mkdir -p scripts

RUN git clone https://github.com/waltforme/kube-bind.git && \
    pushd kube-bind && \
    git checkout origin/kubestellar && \
    popd

ENV PATH=$PATH:/root/go/bin

ADD cmd/             cmd/
ADD config/          config/
ADD hack/            hack/
ADD monitoring/      monitoring/
ADD pkg/             pkg/
ADD scripts/inner/   scripts/inner/
ADD scripts/overlap/ scripts/overlap/
ADD space-framework/ space-framework/
ADD test/            test/
ADD .git/            .git/
ADD .gitattributes Makefile Makefile.venv go.mod go.sum .

RUN make innerbuild GOOS=$TARGETOS GOARCH=$TARGETARCH GIT_DIRTY=$GIT_DIRTY IGNORE_GO_VERSION=yesplease

FROM ghcr.io/waltforme/kube-bind/example-backend:latest AS example-backend-binary

FROM ghcr.io/waltforme/kube-bind/konnector:latest AS konnector-binary

FROM ghcr.io/waltforme/kube-bind/kubectl-bind:latest AS kubectl-bind-binary

FROM ghcr.io/dexidp/dex:v2.37.0 AS dex-binary

FROM redhat/ubi9

WORKDIR /home/kubestellar

RUN dnf install -y jq procps && \
    dnf -y upgrade openssl && \
    groupadd kubestellar && \
    adduser -g kubestellar kubestellar && \
    mkdir -p .kcp

# copy binaries, dex configurations, and kube-bind CRDs from the builder image
COPY --from=builder /home/kubestellar/easy-rsa                           easy-rsa/
COPY --from=builder /root/go/bin                                         /usr/local/bin/
COPY --from=builder /usr/local/bin/kubectl                               /usr/local/bin/kubectl
COPY --from=builder /home/kubestellar/kcp/bin                            kcp/bin/
COPY --from=builder /home/kubestellar/kcp-plugins/bin                    kcp/bin/
COPY --from=builder /home/kubestellar/bin                                bin/
COPY --from=builder /home/kubestellar/config                             config/
COPY --from=builder /home/kubestellar/kube-bind/hack/dex-config-dev.yaml kube-bind/hack/dex-config-dev.yaml
COPY --from=builder /home/kubestellar/kube-bind/deploy/crd               kube-bind/deploy/crd

# copy kube-bind binaries from the kube-bind images
COPY --from=example-backend-binary /ko-app/example-backend kube-bind/bin/
COPY --from=konnector-binary       /ko-app/konnector       kube-bind/bin/
COPY --from=kubectl-bind-binary    /ko-app/kubectl-bind    kube-bind/bin/

# copy binary from the dex-binary image
COPY --from=dex-binary /usr/local/bin/dex dex/bin/dex

# add entry script
ADD core-container/entry.sh entry.sh

RUN chown -R kubestellar:0 /home/kubestellar && \
    chmod -R g=u /home/kubestellar

# setup the environment variables
ENV PATH=/home/kubestellar/bin:/home/kubestellar/kcp/bin:/home/kubestellar/kube-bind/bin:/home/kubestellar/dex/bin:/home/kubestellar/easy-rsa:$PATH
ENV EXTERNAL_HOSTNAME=""
ENV EXTERNAL_PORT=""

# Switch the user
USER kubestellar

# start KubeStellar
CMD [ "/home/kubestellar/entry.sh" ]