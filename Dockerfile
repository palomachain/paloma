###########################
####     Base image    ####
###########################
FROM golang:1.21-buster AS base

# TODO add non-root user

MAINTAINER Matija Martinic <matija@volume.finance>
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

#################################
####    Local chain setup    ####
#################################
FROM ubuntu AS setup-chain-locally
RUN apt-get update && \
	apt-get install -y jq
COPY --from=builder /palomad /palomad
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
