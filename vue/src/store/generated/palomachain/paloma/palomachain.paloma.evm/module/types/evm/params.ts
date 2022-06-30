/* eslint-disable */
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "palomachain.paloma.evm";

/** Params defines the parameters for the module. */
export interface Params {
  chains: Chain[];
}

export interface Chain {
  chainID: string;
  turnstoneID: string;
}

const baseParams: object = {};

export const Params = {
  encode(message: Params, writer: Writer = Writer.create()): Writer {
    for (const v of message.chains) {
      Chain.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Params {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseParams } as Params;
    message.chains = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.chains.push(Chain.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Params {
    const message = { ...baseParams } as Params;
    message.chains = [];
    if (object.chains !== undefined && object.chains !== null) {
      for (const e of object.chains) {
        message.chains.push(Chain.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: Params): unknown {
    const obj: any = {};
    if (message.chains) {
      obj.chains = message.chains.map((e) => (e ? Chain.toJSON(e) : undefined));
    } else {
      obj.chains = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<Params>): Params {
    const message = { ...baseParams } as Params;
    message.chains = [];
    if (object.chains !== undefined && object.chains !== null) {
      for (const e of object.chains) {
        message.chains.push(Chain.fromPartial(e));
      }
    }
    return message;
  },
};

const baseChain: object = { chainID: "", turnstoneID: "" };

export const Chain = {
  encode(message: Chain, writer: Writer = Writer.create()): Writer {
    if (message.chainID !== "") {
      writer.uint32(10).string(message.chainID);
    }
    if (message.turnstoneID !== "") {
      writer.uint32(18).string(message.turnstoneID);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Chain {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseChain } as Chain;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.chainID = reader.string();
          break;
        case 2:
          message.turnstoneID = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Chain {
    const message = { ...baseChain } as Chain;
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = String(object.chainID);
    } else {
      message.chainID = "";
    }
    if (object.turnstoneID !== undefined && object.turnstoneID !== null) {
      message.turnstoneID = String(object.turnstoneID);
    } else {
      message.turnstoneID = "";
    }
    return message;
  },

  toJSON(message: Chain): unknown {
    const obj: any = {};
    message.chainID !== undefined && (obj.chainID = message.chainID);
    message.turnstoneID !== undefined &&
      (obj.turnstoneID = message.turnstoneID);
    return obj;
  },

  fromPartial(object: DeepPartial<Chain>): Chain {
    const message = { ...baseChain } as Chain;
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = object.chainID;
    } else {
      message.chainID = "";
    }
    if (object.turnstoneID !== undefined && object.turnstoneID !== null) {
      message.turnstoneID = object.turnstoneID;
    } else {
      message.turnstoneID = "";
    }
    return message;
  },
};

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
