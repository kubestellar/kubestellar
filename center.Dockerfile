###############################################################################
# Builder image
###############################################################################
FROM redhat/ubi9 AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM

RUN groupadd kubestellar && useradd -g kubestellar kubestellar

WORKDIR /home/kubestellar

RUN mkdir -p .kcp kubestellar-logs && \
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
    mkdir -p bin

ENV PATH=$PATH:/root/go/bin

ADD cmd/		cmd/
ADD config/		config/
ADD hack/		hack/
ADD monitoring/		monitoring/
ADD pkg/		pkg/
ADD scripts/		scripts/
ADD space-provider/	space-provider/
ADD test/		test/
ADD .git/		.git/
ADD .gitattributes Makefile Makefile.venv go.mod go.sum .

# Avoid self-reference
RUN rm scripts/kubectl-kubestellar-deploy

RUN make build

FROM redhat/ubi9

RUN dnf install -y jq procps && \
    dnf -y upgrade openssl && \
    groupadd kubestellar && \
    adduser -g kubestellar kubestellar

WORKDIR /home/kubestellar

# copy binaries from the builder image
COPY --from=builder /home/kubestellar/easy-rsa		easy-rsa/
COPY --from=builder /root/go/bin			/usr/local/bin/
COPY --from=builder /usr/local/bin/kubectl		/usr/local/bin/kubectl
COPY --from=builder /home/kubestellar/kcp/bin        	kcp/bin/
COPY --from=builder /home/kubestellar/kcp-plugins/bin	kcp/bin/
COPY --from=builder /home/kubestellar/bin	      	bin/
COPY --from=builder /home/kubestellar/config	      	config/

# add entry script
ADD user/container/entry.sh entry.sh

RUN chown -R kubestellar:0 /home/kubestellar && \
    chmod -R g=u /home/kubestellar

# setup the environment variables
ENV PATH=/home/kubestellar/bin:/home/kubestellar/kcp/bin:/home/kubestellar/easy-rsa:$PATH
ENV KUBECONFIG=/home/kubestellar/.kcp/admin.kubeconfig
ENV EXTERNAL_HOSTNAME=""
ENV EXTERNAL_PORT=""

# Switch the user
USER kubestellar

# start KubeStellar
CMD [ "/home/kubestellar/entry.sh" ]
