/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "palomachain.paloma.evm";

export interface Valset {
  /** hex addresses on the EVM network */
  validators: string[];
  powers: number[];
  valsetID: number;
}

export interface SubmitLogicCall {
  hexContractAddress: string;
  abi: Uint8Array;
  payload: Uint8Array;
  deadline: number;
}

export interface UpdateValset {
  valset: Valset | undefined;
}

export interface UploadSmartContract {
  bytecode: Uint8Array;
  abi: Uint8Array;
  constructorInput: Uint8Array;
}

export interface Message {
  turnstoneID: string;
  chainID: string;
  submitLogicCall: SubmitLogicCall | undefined;
  updateValset: UpdateValset | undefined;
  uploadSmartContract: UploadSmartContract | undefined;
  compassAddr: string;
}

const baseValset: object = { validators: "", powers: 0, valsetID: 0 };

export const Valset = {
  encode(message: Valset, writer: Writer = Writer.create()): Writer {
    for (const v of message.validators) {
      writer.uint32(10).string(v!);
    }
    writer.uint32(18).fork();
    for (const v of message.powers) {
      writer.uint64(v);
    }
    writer.ldelim();
    if (message.valsetID !== 0) {
      writer.uint32(24).uint64(message.valsetID);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Valset {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseValset } as Valset;
    message.validators = [];
    message.powers = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.validators.push(reader.string());
          break;
        case 2:
          if ((tag & 7) === 2) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.powers.push(longToNumber(reader.uint64() as Long));
            }
          } else {
            message.powers.push(longToNumber(reader.uint64() as Long));
          }
          break;
        case 3:
          message.valsetID = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Valset {
    const message = { ...baseValset } as Valset;
    message.validators = [];
    message.powers = [];
    if (object.validators !== undefined && object.validators !== null) {
      for (const e of object.validators) {
        message.validators.push(String(e));
      }
    }
    if (object.powers !== undefined && object.powers !== null) {
      for (const e of object.powers) {
        message.powers.push(Number(e));
      }
    }
    if (object.valsetID !== undefined && object.valsetID !== null) {
      message.valsetID = Number(object.valsetID);
    } else {
      message.valsetID = 0;
    }
    return message;
  },

  toJSON(message: Valset): unknown {
    const obj: any = {};
    if (message.validators) {
      obj.validators = message.validators.map((e) => e);
    } else {
      obj.validators = [];
    }
    if (message.powers) {
      obj.powers = message.powers.map((e) => e);
    } else {
      obj.powers = [];
    }
    message.valsetID !== undefined && (obj.valsetID = message.valsetID);
    return obj;
  },

  fromPartial(object: DeepPartial<Valset>): Valset {
    const message = { ...baseValset } as Valset;
    message.validators = [];
    message.powers = [];
    if (object.validators !== undefined && object.validators !== null) {
      for (const e of object.validators) {
        message.validators.push(e);
      }
    }
    if (object.powers !== undefined && object.powers !== null) {
      for (const e of object.powers) {
        message.powers.push(e);
      }
    }
    if (object.valsetID !== undefined && object.valsetID !== null) {
      message.valsetID = object.valsetID;
    } else {
      message.valsetID = 0;
    }
    return message;
  },
};

const baseSubmitLogicCall: object = { hexContractAddress: "", deadline: 0 };

export const SubmitLogicCall = {
  encode(message: SubmitLogicCall, writer: Writer = Writer.create()): Writer {
    if (message.hexContractAddress !== "") {
      writer.uint32(10).string(message.hexContractAddress);
    }
    if (message.abi.length !== 0) {
      writer.uint32(18).bytes(message.abi);
    }
    if (message.payload.length !== 0) {
      writer.uint32(26).bytes(message.payload);
    }
    if (message.deadline !== 0) {
      writer.uint32(32).int64(message.deadline);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): SubmitLogicCall {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseSubmitLogicCall } as SubmitLogicCall;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.hexContractAddress = reader.string();
          break;
        case 2:
          message.abi = reader.bytes();
          break;
        case 3:
          message.payload = reader.bytes();
          break;
        case 4:
          message.deadline = longToNumber(reader.int64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): SubmitLogicCall {
    const message = { ...baseSubmitLogicCall } as SubmitLogicCall;
    if (
      object.hexContractAddress !== undefined &&
      object.hexContractAddress !== null
    ) {
      message.hexContractAddress = String(object.hexContractAddress);
    } else {
      message.hexContractAddress = "";
    }
    if (object.abi !== undefined && object.abi !== null) {
      message.abi = bytesFromBase64(object.abi);
    }
    if (object.payload !== undefined && object.payload !== null) {
      message.payload = bytesFromBase64(object.payload);
    }
    if (object.deadline !== undefined && object.deadline !== null) {
      message.deadline = Number(object.deadline);
    } else {
      message.deadline = 0;
    }
    return message;
  },

  toJSON(message: SubmitLogicCall): unknown {
    const obj: any = {};
    message.hexContractAddress !== undefined &&
      (obj.hexContractAddress = message.hexContractAddress);
    message.abi !== undefined &&
      (obj.abi = base64FromBytes(
        message.abi !== undefined ? message.abi : new Uint8Array()
      ));
    message.payload !== undefined &&
      (obj.payload = base64FromBytes(
        message.payload !== undefined ? message.payload : new Uint8Array()
      ));
    message.deadline !== undefined && (obj.deadline = message.deadline);
    return obj;
  },

  fromPartial(object: DeepPartial<SubmitLogicCall>): SubmitLogicCall {
    const message = { ...baseSubmitLogicCall } as SubmitLogicCall;
    if (
      object.hexContractAddress !== undefined &&
      object.hexContractAddress !== null
    ) {
      message.hexContractAddress = object.hexContractAddress;
    } else {
      message.hexContractAddress = "";
    }
    if (object.abi !== undefined && object.abi !== null) {
      message.abi = object.abi;
    } else {
      message.abi = new Uint8Array();
    }
    if (object.payload !== undefined && object.payload !== null) {
      message.payload = object.payload;
    } else {
      message.payload = new Uint8Array();
    }
    if (object.deadline !== undefined && object.deadline !== null) {
      message.deadline = object.deadline;
    } else {
      message.deadline = 0;
    }
    return message;
  },
};

const baseUpdateValset: object = {};

export const UpdateValset = {
  encode(message: UpdateValset, writer: Writer = Writer.create()): Writer {
    if (message.valset !== undefined) {
      Valset.encode(message.valset, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): UpdateValset {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseUpdateValset } as UpdateValset;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.valset = Valset.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): UpdateValset {
    const message = { ...baseUpdateValset } as UpdateValset;
    if (object.valset !== undefined && object.valset !== null) {
      message.valset = Valset.fromJSON(object.valset);
    } else {
      message.valset = undefined;
    }
    return message;
  },

  toJSON(message: UpdateValset): unknown {
    const obj: any = {};
    message.valset !== undefined &&
      (obj.valset = message.valset ? Valset.toJSON(message.valset) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<UpdateValset>): UpdateValset {
    const message = { ...baseUpdateValset } as UpdateValset;
    if (object.valset !== undefined && object.valset !== null) {
      message.valset = Valset.fromPartial(object.valset);
    } else {
      message.valset = undefined;
    }
    return message;
  },
};

const baseUploadSmartContract: object = {};

export const UploadSmartContract = {
  encode(
    message: UploadSmartContract,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.bytecode.length !== 0) {
      writer.uint32(10).bytes(message.bytecode);
    }
    if (message.abi.length !== 0) {
      writer.uint32(18).bytes(message.abi);
    }
    if (message.constructorInput.length !== 0) {
      writer.uint32(26).bytes(message.constructorInput);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): UploadSmartContract {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseUploadSmartContract } as UploadSmartContract;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.bytecode = reader.bytes();
          break;
        case 2:
          message.abi = reader.bytes();
          break;
        case 3:
          message.constructorInput = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): UploadSmartContract {
    const message = { ...baseUploadSmartContract } as UploadSmartContract;
    if (object.bytecode !== undefined && object.bytecode !== null) {
      message.bytecode = bytesFromBase64(object.bytecode);
    }
    if (object.abi !== undefined && object.abi !== null) {
      message.abi = bytesFromBase64(object.abi);
    }
    if (
      object.constructorInput !== undefined &&
      object.constructorInput !== null
    ) {
      message.constructorInput = bytesFromBase64(object.constructorInput);
    }
    return message;
  },

  toJSON(message: UploadSmartContract): unknown {
    const obj: any = {};
    message.bytecode !== undefined &&
      (obj.bytecode = base64FromBytes(
        message.bytecode !== undefined ? message.bytecode : new Uint8Array()
      ));
    message.abi !== undefined &&
      (obj.abi = base64FromBytes(
        message.abi !== undefined ? message.abi : new Uint8Array()
      ));
    message.constructorInput !== undefined &&
      (obj.constructorInput = base64FromBytes(
        message.constructorInput !== undefined
          ? message.constructorInput
          : new Uint8Array()
      ));
    return obj;
  },

  fromPartial(object: DeepPartial<UploadSmartContract>): UploadSmartContract {
    const message = { ...baseUploadSmartContract } as UploadSmartContract;
    if (object.bytecode !== undefined && object.bytecode !== null) {
      message.bytecode = object.bytecode;
    } else {
      message.bytecode = new Uint8Array();
    }
    if (object.abi !== undefined && object.abi !== null) {
      message.abi = object.abi;
    } else {
      message.abi = new Uint8Array();
    }
    if (
      object.constructorInput !== undefined &&
      object.constructorInput !== null
    ) {
      message.constructorInput = object.constructorInput;
    } else {
      message.constructorInput = new Uint8Array();
    }
    return message;
  },
};

const baseMessage: object = { turnstoneID: "", chainID: "", compassAddr: "" };

export const Message = {
  encode(message: Message, writer: Writer = Writer.create()): Writer {
    if (message.turnstoneID !== "") {
      writer.uint32(10).string(message.turnstoneID);
    }
    if (message.chainID !== "") {
      writer.uint32(18).string(message.chainID);
    }
    if (message.submitLogicCall !== undefined) {
      SubmitLogicCall.encode(
        message.submitLogicCall,
        writer.uint32(26).fork()
      ).ldelim();
    }
    if (message.updateValset !== undefined) {
      UpdateValset.encode(
        message.updateValset,
        writer.uint32(34).fork()
      ).ldelim();
    }
    if (message.uploadSmartContract !== undefined) {
      UploadSmartContract.encode(
        message.uploadSmartContract,
        writer.uint32(42).fork()
      ).ldelim();
    }
    if (message.compassAddr !== "") {
      writer.uint32(50).string(message.compassAddr);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Message {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMessage } as Message;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.turnstoneID = reader.string();
          break;
        case 2:
          message.chainID = reader.string();
          break;
        case 3:
          message.submitLogicCall = SubmitLogicCall.decode(
            reader,
            reader.uint32()
          );
          break;
        case 4:
          message.updateValset = UpdateValset.decode(reader, reader.uint32());
          break;
        case 5:
          message.uploadSmartContract = UploadSmartContract.decode(
            reader,
            reader.uint32()
          );
          break;
        case 6:
          message.compassAddr = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Message {
    const message = { ...baseMessage } as Message;
    if (object.turnstoneID !== undefined && object.turnstoneID !== null) {
      message.turnstoneID = String(object.turnstoneID);
    } else {
      message.turnstoneID = "";
    }
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = String(object.chainID);
    } else {
      message.chainID = "";
    }
    if (
      object.submitLogicCall !== undefined &&
      object.submitLogicCall !== null
    ) {
      message.submitLogicCall = SubmitLogicCall.fromJSON(
        object.submitLogicCall
      );
    } else {
      message.submitLogicCall = undefined;
    }
    if (object.updateValset !== undefined && object.updateValset !== null) {
      message.updateValset = UpdateValset.fromJSON(object.updateValset);
    } else {
      message.updateValset = undefined;
    }
    if (
      object.uploadSmartContract !== undefined &&
      object.uploadSmartContract !== null
    ) {
      message.uploadSmartContract = UploadSmartContract.fromJSON(
        object.uploadSmartContract
      );
    } else {
      message.uploadSmartContract = undefined;
    }
    if (object.compassAddr !== undefined && object.compassAddr !== null) {
      message.compassAddr = String(object.compassAddr);
    } else {
      message.compassAddr = "";
    }
    return message;
  },

  toJSON(message: Message): unknown {
    const obj: any = {};
    message.turnstoneID !== undefined &&
      (obj.turnstoneID = message.turnstoneID);
    message.chainID !== undefined && (obj.chainID = message.chainID);
    message.submitLogicCall !== undefined &&
      (obj.submitLogicCall = message.submitLogicCall
        ? SubmitLogicCall.toJSON(message.submitLogicCall)
        : undefined);
    message.updateValset !== undefined &&
      (obj.updateValset = message.updateValset
        ? UpdateValset.toJSON(message.updateValset)
        : undefined);
    message.uploadSmartContract !== undefined &&
      (obj.uploadSmartContract = message.uploadSmartContract
        ? UploadSmartContract.toJSON(message.uploadSmartContract)
        : undefined);
    message.compassAddr !== undefined &&
      (obj.compassAddr = message.compassAddr);
    return obj;
  },

  fromPartial(object: DeepPartial<Message>): Message {
    const message = { ...baseMessage } as Message;
    if (object.turnstoneID !== undefined && object.turnstoneID !== null) {
      message.turnstoneID = object.turnstoneID;
    } else {
      message.turnstoneID = "";
    }
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = object.chainID;
    } else {
      message.chainID = "";
    }
    if (
      object.submitLogicCall !== undefined &&
      object.submitLogicCall !== null
    ) {
      message.submitLogicCall = SubmitLogicCall.fromPartial(
        object.submitLogicCall
      );
    } else {
      message.submitLogicCall = undefined;
    }
    if (object.updateValset !== undefined && object.updateValset !== null) {
      message.updateValset = UpdateValset.fromPartial(object.updateValset);
    } else {
      message.updateValset = undefined;
    }
    if (
      object.uploadSmartContract !== undefined &&
      object.uploadSmartContract !== null
    ) {
      message.uploadSmartContract = UploadSmartContract.fromPartial(
        object.uploadSmartContract
      );
    } else {
      message.uploadSmartContract = undefined;
    }
    if (object.compassAddr !== undefined && object.compassAddr !== null) {
      message.compassAddr = object.compassAddr;
    } else {
      message.compassAddr = "";
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
