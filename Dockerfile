ARG GO_VERSION=1.11.5
FROM golang:${GO_VERSION}-alpine as builder

# Note: make sure to sync any dependencies added here to the "make test" step in the Jenkinsfile.
RUN apk --no-cache add ca-certificates make git openssl-dev libcrypto1.0 gcc libc-dev

WORKDIR /src

COPY go.mod go.sum Makefile ./
RUN make vendor
COPY . .
RUN make all test

FROM alpine:latest

VOLUME  /output \
        /credentials/development \
        /credentials/staging \
        /credentials/production \
        /manifests

RUN apk --no-cache add ca-certificates bash curl openssl-dev libcrypto1.0
COPY --from=0 /src/bin/k8s-secret-projector /bin/k8s-secret-projector

ENV KUBECTL_VERSION=1.9.8 \
    KUBECTL_CHECKSUM=dd7cdde8b7bc4ae74a44bf90f3f0f6e27206787b27a84df62d8421db24f36acd

# install kubectl
RUN curl -L https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl | \
    tee /usr/bin/kubectl | bash -c 'sha256sum -c <(echo -n "${KUBECTL_CHECKSUM}  -" ) && chmod +x /usr/bin/kubectl'

ENTRYPOINT [ \
  "k8s-secret-projector" \
]

CMD [ \
  "--creds-repo=staging=/credentials/staging,production=/credentials/production,development=/credentials/development", \
  "--manifests=/manifests", \
  "--output=/output" \
]
