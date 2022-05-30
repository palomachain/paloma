/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import { Any } from "../google/protobuf/any";

export const protobufPackage = "volumefi.paloma.consensus";

/** message for storing the queued signed message in the internal queue */
export interface QueuedSignedMessage {
  id: number;
  msg: Any | undefined;
  bytesToSign: Uint8Array;
  signData: SignData[];
}

export interface BatchOfConsensusMessages {
  msg: Any | undefined;
}

export interface Batch {
  msgs: Any[];
  bytesToSign: Uint8Array;
}

export interface SignData {
  valAddress: Uint8Array;
  signature: Uint8Array;
  extraData: Uint8Array;
}

const baseQueuedSignedMessage: object = { id: 0 };

export const QueuedSignedMessage = {
  encode(
    message: QueuedSignedMessage,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    if (message.msg !== undefined) {
      Any.encode(message.msg, writer.uint32(18).fork()).ldelim();
    }
    if (message.bytesToSign.length !== 0) {
      writer.uint32(26).bytes(message.bytesToSign);
    }
    for (const v of message.signData) {
      SignData.encode(v!, writer.uint32(34).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueuedSignedMessage {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueuedSignedMessage } as QueuedSignedMessage;
    message.signData = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        case 2:
          message.msg = Any.decode(reader, reader.uint32());
          break;
        case 3:
          message.bytesToSign = reader.bytes();
          break;
        case 4:
          message.signData.push(SignData.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueuedSignedMessage {
    const message = { ...baseQueuedSignedMessage } as QueuedSignedMessage;
    message.signData = [];
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    if (object.msg !== undefined && object.msg !== null) {
      message.msg = Any.fromJSON(object.msg);
    } else {
      message.msg = undefined;
    }
    if (object.bytesToSign !== undefined && object.bytesToSign !== null) {
      message.bytesToSign = bytesFromBase64(object.bytesToSign);
    }
    if (object.signData !== undefined && object.signData !== null) {
      for (const e of object.signData) {
        message.signData.push(SignData.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: QueuedSignedMessage): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    message.msg !== undefined &&
      (obj.msg = message.msg ? Any.toJSON(message.msg) : undefined);
    message.bytesToSign !== undefined &&
      (obj.bytesToSign = base64FromBytes(
        message.bytesToSign !== undefined
          ? message.bytesToSign
          : new Uint8Array()
      ));
    if (message.signData) {
      obj.signData = message.signData.map((e) =>
        e ? SignData.toJSON(e) : undefined
      );
    } else {
      obj.signData = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<QueuedSignedMessage>): QueuedSignedMessage {
    const message = { ...baseQueuedSignedMessage } as QueuedSignedMessage;
    message.signData = [];
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    if (object.msg !== undefined && object.msg !== null) {
      message.msg = Any.fromPartial(object.msg);
    } else {
      message.msg = undefined;
    }
    if (object.bytesToSign !== undefined && object.bytesToSign !== null) {
      message.bytesToSign = object.bytesToSign;
    } else {
      message.bytesToSign = new Uint8Array();
    }
    if (object.signData !== undefined && object.signData !== null) {
      for (const e of object.signData) {
        message.signData.push(SignData.fromPartial(e));
      }
    }
    return message;
  },
};

const baseBatchOfConsensusMessages: object = {};

export const BatchOfConsensusMessages = {
  encode(
    message: BatchOfConsensusMessages,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.msg !== undefined) {
      Any.encode(message.msg, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): BatchOfConsensusMessages {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseBatchOfConsensusMessages,
    } as BatchOfConsensusMessages;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.msg = Any.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): BatchOfConsensusMessages {
    const message = {
      ...baseBatchOfConsensusMessages,
    } as BatchOfConsensusMessages;
    if (object.msg !== undefined && object.msg !== null) {
      message.msg = Any.fromJSON(object.msg);
    } else {
      message.msg = undefined;
    }
    return message;
  },

  toJSON(message: BatchOfConsensusMessages): unknown {
    const obj: any = {};
    message.msg !== undefined &&
      (obj.msg = message.msg ? Any.toJSON(message.msg) : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<BatchOfConsensusMessages>
  ): BatchOfConsensusMessages {
    const message = {
      ...baseBatchOfConsensusMessages,
    } as BatchOfConsensusMessages;
    if (object.msg !== undefined && object.msg !== null) {
      message.msg = Any.fromPartial(object.msg);
    } else {
      message.msg = undefined;
    }
    return message;
  },
};

const baseBatch: object = {};

export const Batch = {
  encode(message: Batch, writer: Writer = Writer.create()): Writer {
    for (const v of message.msgs) {
      Any.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.bytesToSign.length !== 0) {
      writer.uint32(18).bytes(message.bytesToSign);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Batch {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseBatch } as Batch;
    message.msgs = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.msgs.push(Any.decode(reader, reader.uint32()));
          break;
        case 2:
          message.bytesToSign = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Batch {
    const message = { ...baseBatch } as Batch;
    message.msgs = [];
    if (object.msgs !== undefined && object.msgs !== null) {
      for (const e of object.msgs) {
        message.msgs.push(Any.fromJSON(e));
      }
    }
    if (object.bytesToSign !== undefined && object.bytesToSign !== null) {
      message.bytesToSign = bytesFromBase64(object.bytesToSign);
    }
    return message;
  },

  toJSON(message: Batch): unknown {
    const obj: any = {};
    if (message.msgs) {
      obj.msgs = message.msgs.map((e) => (e ? Any.toJSON(e) : undefined));
    } else {
      obj.msgs = [];
    }
    message.bytesToSign !== undefined &&
      (obj.bytesToSign = base64FromBytes(
        message.bytesToSign !== undefined
          ? message.bytesToSign
          : new Uint8Array()
      ));
    return obj;
  },

  fromPartial(object: DeepPartial<Batch>): Batch {
    const message = { ...baseBatch } as Batch;
    message.msgs = [];
    if (object.msgs !== undefined && object.msgs !== null) {
      for (const e of object.msgs) {
        message.msgs.push(Any.fromPartial(e));
      }
    }
    if (object.bytesToSign !== undefined && object.bytesToSign !== null) {
      message.bytesToSign = object.bytesToSign;
    } else {
      message.bytesToSign = new Uint8Array();
    }
    return message;
  },
};

const baseSignData: object = {};

export const SignData = {
  encode(message: SignData, writer: Writer = Writer.create()): Writer {
    if (message.valAddress.length !== 0) {
      writer.uint32(10).bytes(message.valAddress);
    }
    if (message.signature.length !== 0) {
      writer.uint32(18).bytes(message.signature);
    }
    if (message.extraData.length !== 0) {
      writer.uint32(26).bytes(message.extraData);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): SignData {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseSignData } as SignData;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.valAddress = reader.bytes();
          break;
        case 2:
          message.signature = reader.bytes();
          break;
        case 3:
          message.extraData = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): SignData {
    const message = { ...baseSignData } as SignData;
    if (object.valAddress !== undefined && object.valAddress !== null) {
      message.valAddress = bytesFromBase64(object.valAddress);
    }
    if (object.signature !== undefined && object.signature !== null) {
      message.signature = bytesFromBase64(object.signature);
    }
    if (object.extraData !== undefined && object.extraData !== null) {
      message.extraData = bytesFromBase64(object.extraData);
    }
    return message;
  },

  toJSON(message: SignData): unknown {
    const obj: any = {};
    message.valAddress !== undefined &&
      (obj.valAddress = base64FromBytes(
        message.valAddress !== undefined ? message.valAddress : new Uint8Array()
      ));
    message.signature !== undefined &&
      (obj.signature = base64FromBytes(
        message.signature !== undefined ? message.signature : new Uint8Array()
      ));
    message.extraData !== undefined &&
      (obj.extraData = base64FromBytes(
        message.extraData !== undefined ? message.extraData : new Uint8Array()
      ));
    return obj;
  },

  fromPartial(object: DeepPartial<SignData>): SignData {
    const message = { ...baseSignData } as SignData;
    if (object.valAddress !== undefined && object.valAddress !== null) {
      message.valAddress = object.valAddress;
    } else {
      message.valAddress = new Uint8Array();
    }
    if (object.signature !== undefined && object.signature !== null) {
      message.signature = object.signature;
    } else {
      message.signature = new Uint8Array();
    }
    if (object.extraData !== undefined && object.extraData !== null) {
      message.extraData = object.extraData;
    } else {
      message.extraData = new Uint8Array();
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

const atob: (b64: string) => string =
  globalThis.atob ||
  ((b64) => globalThis.Buffer.from(b64, "base64").toString("binary"));
function bytesFromBase64(b64: string): Uint8Array {
  const bin = atob(b64);
  const arr = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; ++i) {
    arr[i] = bin.charCodeAt(i);
  }
  return arr;
}

const btoa: (bin: string) => string =
  globalThis.btoa ||
  ((bin) => globalThis.Buffer.from(bin, "binary").toString("base64"));
function base64FromBytes(arr: Uint8Array): string {
  const bin: string[] = [];
  for (let i = 0; i < arr.byteLength; ++i) {
    bin.push(String.fromCharCode(arr[i]));
  }
  return btoa(bin.join(""));
}

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
