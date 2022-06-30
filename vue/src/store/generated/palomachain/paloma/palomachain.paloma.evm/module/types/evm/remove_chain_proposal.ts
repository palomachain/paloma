/* eslint-disable */
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "palomachain.paloma.evm";

export interface RemoveChainProposal {
  title: string;
  description: string;
  chainID: string;
}

const baseRemoveChainProposal: object = {
  title: "",
  description: "",
  chainID: "",
};

export const RemoveChainProposal = {
  encode(
    message: RemoveChainProposal,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.title !== "") {
      writer.uint32(10).string(message.title);
    }
    if (message.description !== "") {
      writer.uint32(18).string(message.description);
    }
    if (message.chainID !== "") {
      writer.uint32(26).string(message.chainID);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): RemoveChainProposal {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseRemoveChainProposal } as RemoveChainProposal;
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
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): RemoveChainProposal {
    const message = { ...baseRemoveChainProposal } as RemoveChainProposal;
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
    return message;
  },

  toJSON(message: RemoveChainProposal): unknown {
    const obj: any = {};
    message.title !== undefined && (obj.title = message.title);
    message.description !== undefined &&
      (obj.description = message.description);
    message.chainID !== undefined && (obj.chainID = message.chainID);
    return obj;
  },

  fromPartial(object: DeepPartial<RemoveChainProposal>): RemoveChainProposal {
    const message = { ...baseRemoveChainProposal } as RemoveChainProposal;
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
