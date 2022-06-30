/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "palomachain.paloma.evm";

export interface ChainInfo {
  chainID: string;
  smartContractID: string;
  smartContractAddr: string;
  /** used to verify by pigeons if they are at the correct chain */
  referenceBlockHeight: number;
  referenceBlockHash: string;
}

const baseChainInfo: object = {
  chainID: "",
  smartContractID: "",
  smartContractAddr: "",
  referenceBlockHeight: 0,
  referenceBlockHash: "",
};

export const ChainInfo = {
  encode(message: ChainInfo, writer: Writer = Writer.create()): Writer {
    if (message.chainID !== "") {
      writer.uint32(10).string(message.chainID);
    }
    if (message.smartContractID !== "") {
      writer.uint32(18).string(message.smartContractID);
    }
    if (message.smartContractAddr !== "") {
      writer.uint32(26).string(message.smartContractAddr);
    }
    if (message.referenceBlockHeight !== 0) {
      writer.uint32(32).uint64(message.referenceBlockHeight);
    }
    if (message.referenceBlockHash !== "") {
      writer.uint32(42).string(message.referenceBlockHash);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): ChainInfo {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseChainInfo } as ChainInfo;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.chainID = reader.string();
          break;
        case 2:
          message.smartContractID = reader.string();
          break;
        case 3:
          message.smartContractAddr = reader.string();
          break;
        case 4:
          message.referenceBlockHeight = longToNumber(reader.uint64() as Long);
          break;
        case 5:
          message.referenceBlockHash = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): ChainInfo {
    const message = { ...baseChainInfo } as ChainInfo;
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = String(object.chainID);
    } else {
      message.chainID = "";
    }
    if (
      object.smartContractID !== undefined &&
      object.smartContractID !== null
    ) {
      message.smartContractID = String(object.smartContractID);
    } else {
      message.smartContractID = "";
    }
    if (
      object.smartContractAddr !== undefined &&
      object.smartContractAddr !== null
    ) {
      message.smartContractAddr = String(object.smartContractAddr);
    } else {
      message.smartContractAddr = "";
    }
    if (
      object.referenceBlockHeight !== undefined &&
      object.referenceBlockHeight !== null
    ) {
      message.referenceBlockHeight = Number(object.referenceBlockHeight);
    } else {
      message.referenceBlockHeight = 0;
    }
    if (
      object.referenceBlockHash !== undefined &&
      object.referenceBlockHash !== null
    ) {
      message.referenceBlockHash = String(object.referenceBlockHash);
    } else {
      message.referenceBlockHash = "";
    }
    return message;
  },

  toJSON(message: ChainInfo): unknown {
    const obj: any = {};
    message.chainID !== undefined && (obj.chainID = message.chainID);
    message.smartContractID !== undefined &&
      (obj.smartContractID = message.smartContractID);
    message.smartContractAddr !== undefined &&
      (obj.smartContractAddr = message.smartContractAddr);
    message.referenceBlockHeight !== undefined &&
      (obj.referenceBlockHeight = message.referenceBlockHeight);
    message.referenceBlockHash !== undefined &&
      (obj.referenceBlockHash = message.referenceBlockHash);
    return obj;
  },

  fromPartial(object: DeepPartial<ChainInfo>): ChainInfo {
    const message = { ...baseChainInfo } as ChainInfo;
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = object.chainID;
    } else {
      message.chainID = "";
    }
    if (
      object.smartContractID !== undefined &&
      object.smartContractID !== null
    ) {
      message.smartContractID = object.smartContractID;
    } else {
      message.smartContractID = "";
    }
    if (
      object.smartContractAddr !== undefined &&
      object.smartContractAddr !== null
    ) {
      message.smartContractAddr = object.smartContractAddr;
    } else {
      message.smartContractAddr = "";
    }
    if (
      object.referenceBlockHeight !== undefined &&
      object.referenceBlockHeight !== null
    ) {
      message.referenceBlockHeight = object.referenceBlockHeight;
    } else {
      message.referenceBlockHeight = 0;
    }
    if (
      object.referenceBlockHash !== undefined &&
      object.referenceBlockHash !== null
    ) {
      message.referenceBlockHash = object.referenceBlockHash;
    } else {
      message.referenceBlockHash = "";
    }
    return message;
  },
};

declare var self: any | undefined;
declare var window: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

type Builtin = Date | Function | Uint8Array | string | number | undefined;
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

function longToNumber(long: Long): number {
  if (long.gt(Number.MAX_SAFE_INTEGER)) {
    throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
  }
  return long.toNumber();
}

if (util.Long !== Long) {
  util.Long = Long as any;
  configure();
}
