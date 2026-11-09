FROM --platform=$BUILDPLATFORM golang:1.27@sha256:e519d3b0a5b285e5aa53ab9eef02bc6bbb9cb04608389b743a136a9ce75f0d1e AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg pkg
COPY main.go main.go

ARG TARGETOS TARGETARCH
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH
RUN go build -ldflags='-s -w' -o /out/k8s-gen . && \
    go build -ldflags='-s -w' -o /out/docsonnet github.com/jsonnet-libs/docsonnet && \
    go build -ldflags='-s -w' -o /out/jsonnet github.com/google/go-jsonnet/cmd/jsonnet

FROM alpine:3.23@sha256:469c9ea410b9215eaee7976ba6c0639039b444be8e5d638f7abbada189971eb2

WORKDIR /app

RUN apk add --no-cache bash curl git openssh diffutils

COPY --from=mikefarah/yq:4.54.1@sha256:4b3d9475d65571d28cbb19544d3820ec2945e4c8b2f18279394282b8dc3a592e /usr/bin/yq /usr/local/bin/yq
COPY --from=builder /out/ /usr/local/bin/

COPY scripts .
COPY jsonnet jsonnet
COPY LICENSE .

ENTRYPOINT ["./gen.sh"]
