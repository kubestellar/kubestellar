###############################################################################
# Builder image
###############################################################################
FROM redhat/ubi9 AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM

RUN groupadd spacecore && useradd -g spacecore spacecore

WORKDIR /home/spacecore

RUN mkdir -p clusterconfigs && \
    dnf install -y git golang jq procps && \
    go install github.com/mikefarah/yq/v4@v4.34.2 && \
    curl -SL -o /usr/local/bin/kubectl "https://dl.k8s.io/release/v1.25.3/bin/${TARGETPLATFORM}/kubectl" && \
    chmod +x /usr/local/bin/kubectl && \
    git config --global --add safe.directory /home/spacecore && \
    mkdir -p bin

ENV PATH=$PATH:/root/go/bin

ADD cmd/             cmd/
ADD config/          config/
ADD hack/            hack/
ADD pkg/             pkg/
ADD space-provider/  space-provider/
ADD scripts/         scripts/
ADD test/            test/
ADD Makefile go.mod go.sum git-info.txt .

# Avoid self-reference
#RUN rm scripts/kubectl-kubestellar-deploy

RUN make IMPORT_TAGS=yes build

FROM redhat/ubi9

WORKDIR /home/spacecore

RUN dnf install -y jq procps && \
    dnf -y upgrade openssl && \
    groupadd spacecore && \
    adduser -g spacecore spacecore && \
    mkdir -p clusterconfigs

# copy binaries from the builder image
COPY --from=builder /root/go/bin                      /usr/local/bin/
COPY --from=builder /usr/local/bin/kubectl            /usr/local/bin/kubectl
COPY --from=builder /home/spacecore/bin             bin/
COPY --from=builder /home/spacecore/config          config/

# add entry script
ADD entry.sh entry.sh

RUN chown -R spacecore:0 /home/spacecore && \
    chmod -R g=u /home/spacecore

# setup the environment variables
ENV PATH=/home/spacecore/bin:$PATH
ENV KUBECONFIG=/home/spacecore/clusterconfigs/cluster.kubeconfig
ENV EXTERNAL_HOSTNAME=""
ENV EXTERNAL_PORT=""

# Switch the user
USER spacecore

# start KubeStellar
CMD [ "/home/spacecore/entry.sh" ]
