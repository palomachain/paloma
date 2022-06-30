/* eslint-disable */
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "palomachain.paloma.consensus";

export interface SimpleMessage {
  sender: string;
  hello: string;
  world: string;
}

export interface EvenSimplerMessage {
  sender: string;
  boo: string;
}

const baseSimpleMessage: object = { sender: "", hello: "", world: "" };

export const SimpleMessage = {
  encode(message: SimpleMessage, writer: Writer = Writer.create()): Writer {
    if (message.sender !== "") {
      writer.uint32(10).string(message.sender);
    }
    if (message.hello !== "") {
      writer.uint32(18).string(message.hello);
    }
    if (message.world !== "") {
      writer.uint32(26).string(message.world);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): SimpleMessage {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseSimpleMessage } as SimpleMessage;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.string();
          break;
        case 2:
          message.hello = reader.string();
          break;
        case 3:
          message.world = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): SimpleMessage {
    const message = { ...baseSimpleMessage } as SimpleMessage;
    if (object.sender !== undefined && object.sender !== null) {
      message.sender = String(object.sender);
    } else {
      message.sender = "";
    }
    if (object.hello !== undefined && object.hello !== null) {
      message.hello = String(object.hello);
    } else {
      message.hello = "";
    }
    if (object.world !== undefined && object.world !== null) {
      message.world = String(object.world);
    } else {
      message.world = "";
    }
    return message;
  },

  toJSON(message: SimpleMessage): unknown {
    const obj: any = {};
    message.sender !== undefined && (obj.sender = message.sender);
    message.hello !== undefined && (obj.hello = message.hello);
    message.world !== undefined && (obj.world = message.world);
    return obj;
  },

  fromPartial(object: DeepPartial<SimpleMessage>): SimpleMessage {
    const message = { ...baseSimpleMessage } as SimpleMessage;
    if (object.sender !== undefined && object.sender !== null) {
      message.sender = object.sender;
    } else {
      message.sender = "";
    }
    if (object.hello !== undefined && object.hello !== null) {
      message.hello = object.hello;
    } else {
      message.hello = "";
    }
    if (object.world !== undefined && object.world !== null) {
      message.world = object.world;
    } else {
      message.world = "";
    }
    return message;
  },
};

const baseEvenSimplerMessage: object = { sender: "", boo: "" };

export const EvenSimplerMessage = {
  encode(
    message: EvenSimplerMessage,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.sender !== "") {
      writer.uint32(10).string(message.sender);
    }
    if (message.boo !== "") {
      writer.uint32(18).string(message.boo);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): EvenSimplerMessage {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseEvenSimplerMessage } as EvenSimplerMessage;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.string();
          break;
        case 2:
          message.boo = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): EvenSimplerMessage {
    const message = { ...baseEvenSimplerMessage } as EvenSimplerMessage;
    if (object.sender !== undefined && object.sender !== null) {
      message.sender = String(object.sender);
    } else {
      message.sender = "";
    }
    if (object.boo !== undefined && object.boo !== null) {
      message.boo = String(object.boo);
    } else {
      message.boo = "";
    }
    return message;
  },

  toJSON(message: EvenSimplerMessage): unknown {
    const obj: any = {};
    message.sender !== undefined && (obj.sender = message.sender);
    message.boo !== undefined && (obj.boo = message.boo);
    return obj;
  },

  fromPartial(object: DeepPartial<EvenSimplerMessage>): EvenSimplerMessage {
    const message = { ...baseEvenSimplerMessage } as EvenSimplerMessage;
    if (object.sender !== undefined && object.sender !== null) {
      message.sender = object.sender;
    } else {
      message.sender = "";
    }
    if (object.boo !== undefined && object.boo !== null) {
      message.boo = object.boo;
    } else {
      message.boo = "";
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
