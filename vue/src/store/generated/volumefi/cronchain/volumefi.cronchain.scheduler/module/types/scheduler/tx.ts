/* eslint-disable */
import { Reader, Writer } from "protobufjs/minimal";

export const protobufPackage = "volumefi.cronchain.scheduler";

export interface MsgSigningQueueMessage {
  creator: string;
  id: string;
}

export interface MsgSigningQueueMessageResponse {}

const baseMsgSigningQueueMessage: object = { creator: "", id: "" };

export const MsgSigningQueueMessage = {
  encode(
    message: MsgSigningQueueMessage,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.id !== "") {
      writer.uint32(18).string(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgSigningQueueMessage {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgSigningQueueMessage } as MsgSigningQueueMessage;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.id = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgSigningQueueMessage {
    const message = { ...baseMsgSigningQueueMessage } as MsgSigningQueueMessage;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = String(object.id);
    } else {
      message.id = "";
    }
    return message;
  },

  toJSON(message: MsgSigningQueueMessage): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgSigningQueueMessage>
  ): MsgSigningQueueMessage {
    const message = { ...baseMsgSigningQueueMessage } as MsgSigningQueueMessage;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = "";
    }
    return message;
  },
};

const baseMsgSigningQueueMessageResponse: object = {};

export const MsgSigningQueueMessageResponse = {
  encode(
    _: MsgSigningQueueMessageResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgSigningQueueMessageResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgSigningQueueMessageResponse,
    } as MsgSigningQueueMessageResponse;
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

  fromJSON(_: any): MsgSigningQueueMessageResponse {
    const message = {
      ...baseMsgSigningQueueMessageResponse,
    } as MsgSigningQueueMessageResponse;
    return message;
  },

  toJSON(_: MsgSigningQueueMessageResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgSigningQueueMessageResponse>
  ): MsgSigningQueueMessageResponse {
    const message = {
      ...baseMsgSigningQueueMessageResponse,
    } as MsgSigningQueueMessageResponse;
    return message;
  },
};

/** Msg defines the Msg service. */
export interface Msg {
  /** this line is used by starport scaffolding # proto/tx/rpc */
  SigningQueueMessage(
    request: MsgSigningQueueMessage
  ): Promise<MsgSigningQueueMessageResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  SigningQueueMessage(
    request: MsgSigningQueueMessage
  ): Promise<MsgSigningQueueMessageResponse> {
    const data = MsgSigningQueueMessage.encode(request).finish();
    const promise = this.rpc.request(
      "volumefi.cronchain.scheduler.Msg",
      "SigningQueueMessage",
      data
    );
    return promise.then((data) =>
      MsgSigningQueueMessageResponse.decode(new Reader(data))
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
