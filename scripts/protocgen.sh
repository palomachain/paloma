#!/usr/bin/env bash

set -eo pipefail

protoc_gen_gocosmos() {
  if ! grep "github.com/gogo/protobuf => github.com/regen-network/protobuf" go.mod &>/dev/null ; then
    echo -e "\tPlease run this command from somewhere inside the cosmos-sdk folder."
    return 1
  fi

  # get protoc executions
  go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@latest 2>/dev/null
  
  # get cosmos sdk from github
  go get github.com/cosmos/cosmos-sdk@v0.45.10 2>/dev/null
}

protoc_gen_gocosmos

# get the path of the cosmos-sdk repo from go/pkg/mod
cosmos_sdk_dir=$(go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk)
proto_dirs=$(find . -path ./third_party -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)

for dir in $proto_dirs; do
  # generate protobuf bind
  buf protoc \
    -I "proto" \
    -I "$cosmos_sdk_dir/third_party/proto" \
    -I "$cosmos_sdk_dir/proto" \
    --gocosmos_out=plugins=interfacetype+grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
    $(find "${dir}" -name '*.proto')

  # generate grpc gateway
  buf protoc \
    -I "proto" \
    -I "$cosmos_sdk_dir/third_party/proto" \
    -I "$cosmos_sdk_dir/proto" \
    --grpc-gateway_out=logtostderr=true:. \
    $(find "${dir}" -maxdepth 1 -name '*.proto')
done

cp -r ./github.com/palomachain/paloma/* ./
rm -rf ./github.com

# go mod tidy -compat=1.18
