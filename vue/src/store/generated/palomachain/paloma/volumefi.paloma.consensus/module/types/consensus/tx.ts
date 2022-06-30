/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";

export const protobufPackage = "volumefi.paloma.consensus";

export interface MsgAddMessagesSignatures {
  creator: string;
  signedMessages: MsgAddMessagesSignatures_MsgSignedMessage[];
}

export interface MsgAddMessagesSignatures_MsgSignedMessage {
  id: number;
  queueTypeName: string;
  signature: Uint8Array;
  extraData: Uint8Array;
  signedByAddress: string;
}

export interface MsgAddMessagesSignaturesResponse {}

export interface MsgDeleteJob {
  creator: string;
  queueTypeName: string;
  messageID: number;
}

export interface MsgDeleteJobResponse {}

const baseMsgAddMessagesSignatures: object = { creator: "" };

export const MsgAddMessagesSignatures = {
  encode(
    message: MsgAddMessagesSignatures,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    for (const v of message.signedMessages) {
      MsgAddMessagesSignatures_MsgSignedMessage.encode(
        v!,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgAddMessagesSignatures {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgAddMessagesSignatures,
    } as MsgAddMessagesSignatures;
    message.signedMessages = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.signedMessages.push(
            MsgAddMessagesSignatures_MsgSignedMessage.decode(
              reader,
              reader.uint32()
            )
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgAddMessagesSignatures {
    const message = {
      ...baseMsgAddMessagesSignatures,
    } as MsgAddMessagesSignatures;
    message.signedMessages = [];
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.signedMessages !== undefined && object.signedMessages !== null) {
      for (const e of object.signedMessages) {
        message.signedMessages.push(
          MsgAddMessagesSignatures_MsgSignedMessage.fromJSON(e)
        );
      }
    }
    return message;
  },

  toJSON(message: MsgAddMessagesSignatures): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    if (message.signedMessages) {
      obj.signedMessages = message.signedMessages.map((e) =>
        e ? MsgAddMessagesSignatures_MsgSignedMessage.toJSON(e) : undefined
      );
    } else {
      obj.signedMessages = [];
    }
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgAddMessagesSignatures>
  ): MsgAddMessagesSignatures {
    const message = {
      ...baseMsgAddMessagesSignatures,
    } as MsgAddMessagesSignatures;
    message.signedMessages = [];
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.signedMessages !== undefined && object.signedMessages !== null) {
      for (const e of object.signedMessages) {
        message.signedMessages.push(
          MsgAddMessagesSignatures_MsgSignedMessage.fromPartial(e)
        );
      }
    }
    return message;
  },
};

const baseMsgAddMessagesSignatures_MsgSignedMessage: object = {
  id: 0,
  queueTypeName: "",
  signedByAddress: "",
};

export const MsgAddMessagesSignatures_MsgSignedMessage = {
  encode(
    message: MsgAddMessagesSignatures_MsgSignedMessage,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    if (message.queueTypeName !== "") {
      writer.uint32(18).string(message.queueTypeName);
    }
    if (message.signature.length !== 0) {
      writer.uint32(26).bytes(message.signature);
    }
    if (message.extraData.length !== 0) {
      writer.uint32(34).bytes(message.extraData);
    }
    if (message.signedByAddress !== "") {
      writer.uint32(42).string(message.signedByAddress);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgAddMessagesSignatures_MsgSignedMessage {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgAddMessagesSignatures_MsgSignedMessage,
    } as MsgAddMessagesSignatures_MsgSignedMessage;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        case 2:
          message.queueTypeName = reader.string();
          break;
        case 3:
          message.signature = reader.bytes();
          break;
        case 4:
          message.extraData = reader.bytes();
          break;
        case 5:
          message.signedByAddress = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgAddMessagesSignatures_MsgSignedMessage {
    const message = {
      ...baseMsgAddMessagesSignatures_MsgSignedMessage,
    } as MsgAddMessagesSignatures_MsgSignedMessage;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    if (object.queueTypeName !== undefined && object.queueTypeName !== null) {
      message.queueTypeName = String(object.queueTypeName);
    } else {
      message.queueTypeName = "";
    }
    if (object.signature !== undefined && object.signature !== null) {
      message.signature = bytesFromBase64(object.signature);
    }
    if (object.extraData !== undefined && object.extraData !== null) {
      message.extraData = bytesFromBase64(object.extraData);
    }
    if (
      object.signedByAddress !== undefined &&
      object.signedByAddress !== null
    ) {
      message.signedByAddress = String(object.signedByAddress);
    } else {
      message.signedByAddress = "";
    }
    return message;
  },

  toJSON(message: MsgAddMessagesSignatures_MsgSignedMessage): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    message.queueTypeName !== undefined &&
      (obj.queueTypeName = message.queueTypeName);
    message.signature !== undefined &&
      (obj.signature = base64FromBytes(
        message.signature !== undefined ? message.signature : new Uint8Array()
      ));
    message.extraData !== undefined &&
      (obj.extraData = base64FromBytes(
        message.extraData !== undefined ? message.extraData : new Uint8Array()
      ));
    message.signedByAddress !== undefined &&
      (obj.signedByAddress = message.signedByAddress);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgAddMessagesSignatures_MsgSignedMessage>
  ): MsgAddMessagesSignatures_MsgSignedMessage {
    const message = {
      ...baseMsgAddMessagesSignatures_MsgSignedMessage,
    } as MsgAddMessagesSignatures_MsgSignedMessage;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    if (object.queueTypeName !== undefined && object.queueTypeName !== null) {
      message.queueTypeName = object.queueTypeName;
    } else {
      message.queueTypeName = "";
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
    if (
      object.signedByAddress !== undefined &&
      object.signedByAddress !== null
    ) {
      message.signedByAddress = object.signedByAddress;
    } else {
      message.signedByAddress = "";
    }
    return message;
  },
};

const baseMsgAddMessagesSignaturesResponse: object = {};

export const MsgAddMessagesSignaturesResponse = {
  encode(
    _: MsgAddMessagesSignaturesResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgAddMessagesSignaturesResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgAddMessagesSignaturesResponse,
    } as MsgAddMessagesSignaturesResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgAddMessagesSignaturesResponse {
    const message = {
      ...baseMsgAddMessagesSignaturesResponse,
    } as MsgAddMessagesSignaturesResponse;
    return message;
  },

  toJSON(_: MsgAddMessagesSignaturesResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgAddMessagesSignaturesResponse>
  ): MsgAddMessagesSignaturesResponse {
    const message = {
      ...baseMsgAddMessagesSignaturesResponse,
    } as MsgAddMessagesSignaturesResponse;
    return message;
  },
};

const baseMsgDeleteJob: object = {
  creator: "",
  queueTypeName: "",
  messageID: 0,
};

export const MsgDeleteJob = {
  encode(message: MsgDeleteJob, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.queueTypeName !== "") {
      writer.uint32(18).string(message.queueTypeName);
    }
    if (message.messageID !== 0) {
      writer.uint32(24).uint64(message.messageID);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDeleteJob {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDeleteJob } as MsgDeleteJob;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.queueTypeName = reader.string();
          break;
        case 3:
          message.messageID = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgDeleteJob {
    const message = { ...baseMsgDeleteJob } as MsgDeleteJob;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.queueTypeName !== undefined && object.queueTypeName !== null) {
      message.queueTypeName = String(object.queueTypeName);
    } else {
      message.queueTypeName = "";
    }
    if (object.messageID !== undefined && object.messageID !== null) {
      message.messageID = Number(object.messageID);
    } else {
      message.messageID = 0;
    }
    return message;
  },

  toJSON(message: MsgDeleteJob): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.queueTypeName !== undefined &&
      (obj.queueTypeName = message.queueTypeName);
    message.messageID !== undefined && (obj.messageID = message.messageID);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgDeleteJob>): MsgDeleteJob {
    const message = { ...baseMsgDeleteJob } as MsgDeleteJob;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.queueTypeName !== undefined && object.queueTypeName !== null) {
      message.queueTypeName = object.queueTypeName;
    } else {
      message.queueTypeName = "";
    }
    if (object.messageID !== undefined && object.messageID !== null) {
      message.messageID = object.messageID;
    } else {
      message.messageID = 0;
    }
    return message;
  },
};

const baseMsgDeleteJobResponse: object = {};

export const MsgDeleteJobResponse = {
  encode(_: MsgDeleteJobResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDeleteJobResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDeleteJobResponse } as MsgDeleteJobResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgDeleteJobResponse {
    const message = { ...baseMsgDeleteJobResponse } as MsgDeleteJobResponse;
    return message;
  },

  toJSON(_: MsgDeleteJobResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<MsgDeleteJobResponse>): MsgDeleteJobResponse {
    const message = { ...baseMsgDeleteJobResponse } as MsgDeleteJobResponse;
    return message;
  },
};

/** Msg defines the Msg service. */
export interface Msg {
  AddMessagesSignatures(
    request: MsgAddMessagesSignatures
  ): Promise<MsgAddMessagesSignaturesResponse>;
  /** this line is used by starport scaffolding # proto/tx/rpc */
  DeleteJob(request: MsgDeleteJob): Promise<MsgDeleteJobResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  AddMessagesSignatures(
    request: MsgAddMessagesSignatures
  ): Promise<MsgAddMessagesSignaturesResponse> {
    const data = MsgAddMessagesSignatures.encode(request).finish();
    const promise = this.rpc.request(
      "volumefi.paloma.consensus.Msg",
      "AddMessagesSignatures",
      data
    );
    return promise.then((data) =>
      MsgAddMessagesSignaturesResponse.decode(new Reader(data))
    );
  }

  DeleteJob(request: MsgDeleteJob): Promise<MsgDeleteJobResponse> {
    const data = MsgDeleteJob.encode(request).finish();
    const promise = this.rpc.request(
      "volumefi.paloma.consensus.Msg",
      "DeleteJob",
      data
    );
    return promise.then((data) =>
      MsgDeleteJobResponse.decode(new Reader(data))
    );
  }
}

interface Rpc {
  request(
    service: string,
    method: string,
    data: Uint8Array
  ): Promise<Uint8Array>;
}

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
