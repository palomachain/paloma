#!/usr/bin/env bash


# TODO: make this better! 
protoc -I . -I third_party/proto/ --gocosmos_out=plugins=interfacetype+grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. runner/client/terra/proto/tx.proto
protoc -I . -I third_party/proto/ --gocosmos_out=plugins=interfacetype+grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:.  runner/testdata/proto/*
protoc -I . -I third_party/proto/ --gocosmos_out=plugins=interfacetype+grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:.  x/concensus/testdata/proto/*
protoc -I proto -I third_party/proto/ --gocosmos_out=plugins=interfacetype+grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:.  proto/scheduler/*.proto
protoc -I proto -I third_party/proto/ --gocosmos_out=plugins=interfacetype+grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:.  proto/concensus/*.proto


cp -r ./github.com/terra-money/core/x/wasm/types runner/client/terra/types
cp -r ./github.com/volumefi/cronchain/runner/* runner
cp -r ./github.com/volumefi/cronchain/x/* x


rm -rf ./github.com
