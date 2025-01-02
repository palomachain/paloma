###########################
####     Base image    ####
###########################
FROM golang:1.22-buster AS base

# TODO add non-root user
LABEL org.opencontainers.image.authors="christian@volume.finance"
WORKDIR /app

###########################
####     Builder       ####
###########################
FROM base AS builder
COPY . /app
RUN \
  --mount=type=cache,target=/go/pkg/mod \
   --mount=type=cache,target=/root/.cache/go-build \
   cd /app && go build -o /palomad ./cmd/palomad

################################
####     Static builder     ####
################################
FROM golang:1.22-alpine3.19 AS static-builder
WORKDIR /app
COPY . .
RUN apk add --no-cache make git gcc linux-headers build-base curl

RUN go mod download

RUN WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm/v2 | cut -d ' ' -f 2) && \
  wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$(uname -m).a -O /lib/libwasmvm.$(uname -m).a

RUN CGO_ENABLED=1 GOOS=linux go build -o ./build/palomad -tags 'netgo' -ldflags "\
  -X github.com/cosmos/cosmos-sdk/version.Name=paloma \
  -X github.com/cosmos/cosmos-sdk/version.AppName=palomad \
  -X github.com/cosmos/cosmos-sdk/version.Version=$(git describe --tags) \
  -X github.com/cosmos/cosmos-sdk/version.Commit=$(git log -1 --format='%H') \
  -X 'github.com/cosmos/cosmos-sdk/version.BuildTags=netgo' \
  -X github.com/cometbft/cometbft/version.TMCoreSemVer=$(go list -m github.com/cometbft/cometbft | sed 's:.* ::') \
  -linkmode=external -extldflags '-Wl,-z,muldefs -static'" ./cmd/palomad

ENTRYPOINT ["/app/build/palomad"]

#################################
####    Local chain setup    ####
#################################
FROM ubuntu AS setup-chain-locally
RUN apt-get update && \
  apt-get install -y jq
  OPY --from=builder /palomad /palomad
COPY --from=builder /app/scripts/setup-chain-validator.sh /app/scripts/setup-chain-validator.sh
RUN PALOMA_CMD="/palomad" /app/scripts/setup-chain-validator.sh

###########################
#### Local development ####
###########################
FROM base AS local-dev
RUN cd /tmp && go install github.com/cosmtrek/air@latest
COPY --from=setup-chain-locally /root/.paloma /root/.paloma

# air is not set to entrypoint because I want to override that behaviour
# when using docker-compose run.
CMD ["air"]


###########################
####  Local testnet    ####
###########################
FROM ubuntu AS local-testnet
ENTRYPOINT ["/palomad"]
COPY --from=builder /palomad /palomad


###########################
####     Release       ####
###########################
FROM base AS release
RUN go install github.com/goreleaser/goreleaser@latest
COPY . /app

CMD ["goreleaser", "release", "--rm-dist"]
