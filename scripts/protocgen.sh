#!/usr/bin/env bash


# TODO: make this better! 
protoc -I . -I third_party/proto/ --gocosmos_out=plugins=interfacetype+grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. runner/client/terra/proto/tx.proto
protoc -I . -I third_party/proto/ --gocosmos_out=plugins=interfacetype+grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:.  runner/testdata/proto/*


cp -r ./github.com/terra-money/core/x/wasm/types runner/client/terra/types
cp -r ./github.com/volumefi/cronchain/runner/* runner


rm -rf ./github.com
