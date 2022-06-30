/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "palomachain.paloma.evm";

export interface AddChainProposal {
  title: string;
  description: string;
  chainID: string;
  blockHeight: number;
  blockHashAtHeight: string;
}

const baseAddChainProposal: object = {
  title: "",
  description: "",
  chainID: "",
  blockHeight: 0,
  blockHashAtHeight: "",
};

export const AddChainProposal = {
  encode(message: AddChainProposal, writer: Writer = Writer.create()): Writer {
    if (message.title !== "") {
      writer.uint32(10).string(message.title);
    }
    if (message.description !== "") {
      writer.uint32(18).string(message.description);
    }
    if (message.chainID !== "") {
      writer.uint32(26).string(message.chainID);
    }
    if (message.blockHeight !== 0) {
      writer.uint32(32).uint64(message.blockHeight);
    }
    if (message.blockHashAtHeight !== "") {
      writer.uint32(42).string(message.blockHashAtHeight);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): AddChainProposal {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseAddChainProposal } as AddChainProposal;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.title = reader.string();
          break;
        case 2:
          message.description = reader.string();
          break;
        case 3:
          message.chainID = reader.string();
          break;
        case 4:
          message.blockHeight = longToNumber(reader.uint64() as Long);
          break;
        case 5:
          message.blockHashAtHeight = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): AddChainProposal {
    const message = { ...baseAddChainProposal } as AddChainProposal;
    if (object.title !== undefined && object.title !== null) {
      message.title = String(object.title);
    } else {
      message.title = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = String(object.description);
    } else {
      message.description = "";
    }
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = String(object.chainID);
    } else {
      message.chainID = "";
    }
    if (object.blockHeight !== undefined && object.blockHeight !== null) {
      message.blockHeight = Number(object.blockHeight);
    } else {
      message.blockHeight = 0;
    }
    if (
      object.blockHashAtHeight !== undefined &&
      object.blockHashAtHeight !== null
    ) {
      message.blockHashAtHeight = String(object.blockHashAtHeight);
    } else {
      message.blockHashAtHeight = "";
    }
    return message;
  },

  toJSON(message: AddChainProposal): unknown {
    const obj: any = {};
    message.title !== undefined && (obj.title = message.title);
    message.description !== undefined &&
      (obj.description = message.description);
    message.chainID !== undefined && (obj.chainID = message.chainID);
    message.blockHeight !== undefined &&
      (obj.blockHeight = message.blockHeight);
    message.blockHashAtHeight !== undefined &&
      (obj.blockHashAtHeight = message.blockHashAtHeight);
    return obj;
  },

  fromPartial(object: DeepPartial<AddChainProposal>): AddChainProposal {
    const message = { ...baseAddChainProposal } as AddChainProposal;
    if (object.title !== undefined && object.title !== null) {
      message.title = object.title;
    } else {
      message.title = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = object.description;
    } else {
      message.description = "";
    }
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = object.chainID;
    } else {
      message.chainID = "";
    }
    if (object.blockHeight !== undefined && object.blockHeight !== null) {
      message.blockHeight = object.blockHeight;
    } else {
      message.blockHeight = 0;
    }
    if (
      object.blockHashAtHeight !== undefined &&
      object.blockHashAtHeight !== null
    ) {
      message.blockHashAtHeight = object.blockHashAtHeight;
    } else {
      message.blockHashAtHeight = "";
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
