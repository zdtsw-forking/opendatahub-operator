# Build the manager binary
ARG GOLANG_VERSION=1.21
ARG USE_LOCAL=false
ARG OVERWRITE_MANIFESTS=""

################################################################################
FROM registry.access.redhat.com/ubi8/go-toolset:$GOLANG_VERSION as builder_local_false
ARG OVERWRITE_MANIFESTS
# Get all manifests from remote git repo to builder_local_false by script
USER root
WORKDIR /opt
COPY get_all_manifests.sh get_all_manifests.sh
RUN ./get_all_manifests.sh ${OVERWRITE_MANIFESTS}

################################################################################
FROM registry.access.redhat.com/ubi8/go-toolset:$GOLANG_VERSION as builder_local_true
# Get all manifests from local to builder_local_true
USER root
WORKDIR /opt
# copy local manifests to build
COPY odh-manifests/ /opt/odh-manifests/

################################################################################
FROM builder_local_${USE_LOCAL} as builder
USER root
WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY apis/ apis/
COPY components/ components/
COPY controllers/ controllers/
COPY main.go main.go
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 go install github.com/go-delve/delve/cmd/dlv@latest
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -gcflags="all=-N -l" -a -o manager main.go

################################################################################
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
WORKDIR /
COPY --from=builder /usr/bin/dlv .
COPY --from=builder /workspace/manager .
COPY --chown=1001:0 --from=builder /opt/odh-manifests /opt/manifests
# Recursive change all files
RUN chown -R 1001:0 /opt/manifests &&\
   chmod -R a+r /opt/manifests
USER 1001

ENTRYPOINT ["/dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "--log=true", "exec", "/manager", "--"]

#1 kubectl port-forward -n <operatorn_namespace> po/<pod_name> 40000:40000
#then
#1 docker run --name <pod_name> -p 40000:40000 <image_name>
#2 dvl connect localhost:40000

