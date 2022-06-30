/* eslint-disable */
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "palomachain.paloma.evm";

export interface DeployNewSmartContract {
  chainID: string;
  abiJSON: string;
  bytecode: string;
}

const baseDeployNewSmartContract: object = {
  chainID: "",
  abiJSON: "",
  bytecode: "",
};

export const DeployNewSmartContract = {
  encode(
    message: DeployNewSmartContract,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.chainID !== "") {
      writer.uint32(10).string(message.chainID);
    }
    if (message.abiJSON !== "") {
      writer.uint32(18).string(message.abiJSON);
    }
    if (message.bytecode !== "") {
      writer.uint32(26).string(message.bytecode);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): DeployNewSmartContract {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseDeployNewSmartContract } as DeployNewSmartContract;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.chainID = reader.string();
          break;
        case 2:
          message.abiJSON = reader.string();
          break;
        case 3:
          message.bytecode = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): DeployNewSmartContract {
    const message = { ...baseDeployNewSmartContract } as DeployNewSmartContract;
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = String(object.chainID);
    } else {
      message.chainID = "";
    }
    if (object.abiJSON !== undefined && object.abiJSON !== null) {
      message.abiJSON = String(object.abiJSON);
    } else {
      message.abiJSON = "";
    }
    if (object.bytecode !== undefined && object.bytecode !== null) {
      message.bytecode = String(object.bytecode);
    } else {
      message.bytecode = "";
    }
    return message;
  },

  toJSON(message: DeployNewSmartContract): unknown {
    const obj: any = {};
    message.chainID !== undefined && (obj.chainID = message.chainID);
    message.abiJSON !== undefined && (obj.abiJSON = message.abiJSON);
    message.bytecode !== undefined && (obj.bytecode = message.bytecode);
    return obj;
  },

  fromPartial(
    object: DeepPartial<DeployNewSmartContract>
  ): DeployNewSmartContract {
    const message = { ...baseDeployNewSmartContract } as DeployNewSmartContract;
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = object.chainID;
    } else {
      message.chainID = "";
    }
    if (object.abiJSON !== undefined && object.abiJSON !== null) {
      message.abiJSON = object.abiJSON;
    } else {
      message.abiJSON = "";
    }
    if (object.bytecode !== undefined && object.bytecode !== null) {
      message.bytecode = object.bytecode;
    } else {
      message.bytecode = "";
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
