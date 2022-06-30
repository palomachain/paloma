/* eslint-disable */
import { Reader, Writer } from "protobufjs/minimal";

export const protobufPackage = "palomachain.paloma.evm";

export interface MsgSubmitNewJob {
  creator: string;
  hexSmartContractAddress: string;
  hexPayload: string;
  abi: string;
  method: string;
  chainType: string;
  chainID: string;
}

export interface MsgSubmitNewJobResponse {}

export interface MsgUploadNewSmartContractTemp {
  creator: string;
  abi: string;
  bytecode: string;
  constructorInput: string;
  chainID: string;
}

export interface MsgUploadNewSmartContractTempResponse {}

const baseMsgSubmitNewJob: object = {
  creator: "",
  hexSmartContractAddress: "",
  hexPayload: "",
  abi: "",
  method: "",
  chainType: "",
  chainID: "",
};

export const MsgSubmitNewJob = {
  encode(message: MsgSubmitNewJob, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.hexSmartContractAddress !== "") {
      writer.uint32(18).string(message.hexSmartContractAddress);
    }
    if (message.hexPayload !== "") {
      writer.uint32(26).string(message.hexPayload);
    }
    if (message.abi !== "") {
      writer.uint32(34).string(message.abi);
    }
    if (message.method !== "") {
      writer.uint32(42).string(message.method);
    }
    if (message.chainType !== "") {
      writer.uint32(50).string(message.chainType);
    }
    if (message.chainID !== "") {
      writer.uint32(58).string(message.chainID);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgSubmitNewJob {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgSubmitNewJob } as MsgSubmitNewJob;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.hexSmartContractAddress = reader.string();
          break;
        case 3:
          message.hexPayload = reader.string();
          break;
        case 4:
          message.abi = reader.string();
          break;
        case 5:
          message.method = reader.string();
          break;
        case 6:
          message.chainType = reader.string();
          break;
        case 7:
          message.chainID = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgSubmitNewJob {
    const message = { ...baseMsgSubmitNewJob } as MsgSubmitNewJob;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (
      object.hexSmartContractAddress !== undefined &&
      object.hexSmartContractAddress !== null
    ) {
      message.hexSmartContractAddress = String(object.hexSmartContractAddress);
    } else {
      message.hexSmartContractAddress = "";
    }
    if (object.hexPayload !== undefined && object.hexPayload !== null) {
      message.hexPayload = String(object.hexPayload);
    } else {
      message.hexPayload = "";
    }
    if (object.abi !== undefined && object.abi !== null) {
      message.abi = String(object.abi);
    } else {
      message.abi = "";
    }
    if (object.method !== undefined && object.method !== null) {
      message.method = String(object.method);
    } else {
      message.method = "";
    }
    if (object.chainType !== undefined && object.chainType !== null) {
      message.chainType = String(object.chainType);
    } else {
      message.chainType = "";
    }
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = String(object.chainID);
    } else {
      message.chainID = "";
    }
    return message;
  },

  toJSON(message: MsgSubmitNewJob): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.hexSmartContractAddress !== undefined &&
      (obj.hexSmartContractAddress = message.hexSmartContractAddress);
    message.hexPayload !== undefined && (obj.hexPayload = message.hexPayload);
    message.abi !== undefined && (obj.abi = message.abi);
    message.method !== undefined && (obj.method = message.method);
    message.chainType !== undefined && (obj.chainType = message.chainType);
    message.chainID !== undefined && (obj.chainID = message.chainID);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgSubmitNewJob>): MsgSubmitNewJob {
    const message = { ...baseMsgSubmitNewJob } as MsgSubmitNewJob;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (
      object.hexSmartContractAddress !== undefined &&
      object.hexSmartContractAddress !== null
    ) {
      message.hexSmartContractAddress = object.hexSmartContractAddress;
    } else {
      message.hexSmartContractAddress = "";
    }
    if (object.hexPayload !== undefined && object.hexPayload !== null) {
      message.hexPayload = object.hexPayload;
    } else {
      message.hexPayload = "";
    }
    if (object.abi !== undefined && object.abi !== null) {
      message.abi = object.abi;
    } else {
      message.abi = "";
    }
    if (object.method !== undefined && object.method !== null) {
      message.method = object.method;
    } else {
      message.method = "";
    }
    if (object.chainType !== undefined && object.chainType !== null) {
      message.chainType = object.chainType;
    } else {
      message.chainType = "";
    }
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = object.chainID;
    } else {
      message.chainID = "";
    }
    return message;
  },
};

const baseMsgSubmitNewJobResponse: object = {};

export const MsgSubmitNewJobResponse = {
  encode(_: MsgSubmitNewJobResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgSubmitNewJobResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgSubmitNewJobResponse,
    } as MsgSubmitNewJobResponse;
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

  fromJSON(_: any): MsgSubmitNewJobResponse {
    const message = {
      ...baseMsgSubmitNewJobResponse,
    } as MsgSubmitNewJobResponse;
    return message;
  },

  toJSON(_: MsgSubmitNewJobResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgSubmitNewJobResponse>
  ): MsgSubmitNewJobResponse {
    const message = {
      ...baseMsgSubmitNewJobResponse,
    } as MsgSubmitNewJobResponse;
    return message;
  },
};

const baseMsgUploadNewSmartContractTemp: object = {
  creator: "",
  abi: "",
  bytecode: "",
  constructorInput: "",
  chainID: "",
};

export const MsgUploadNewSmartContractTemp = {
  encode(
    message: MsgUploadNewSmartContractTemp,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.abi !== "") {
      writer.uint32(18).string(message.abi);
    }
    if (message.bytecode !== "") {
      writer.uint32(26).string(message.bytecode);
    }
    if (message.constructorInput !== "") {
      writer.uint32(34).string(message.constructorInput);
    }
    if (message.chainID !== "") {
      writer.uint32(42).string(message.chainID);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgUploadNewSmartContractTemp {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgUploadNewSmartContractTemp,
    } as MsgUploadNewSmartContractTemp;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.abi = reader.string();
          break;
        case 3:
          message.bytecode = reader.string();
          break;
        case 4:
          message.constructorInput = reader.string();
          break;
        case 5:
          message.chainID = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgUploadNewSmartContractTemp {
    const message = {
      ...baseMsgUploadNewSmartContractTemp,
    } as MsgUploadNewSmartContractTemp;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.abi !== undefined && object.abi !== null) {
      message.abi = String(object.abi);
    } else {
      message.abi = "";
    }
    if (object.bytecode !== undefined && object.bytecode !== null) {
      message.bytecode = String(object.bytecode);
    } else {
      message.bytecode = "";
    }
    if (
      object.constructorInput !== undefined &&
      object.constructorInput !== null
    ) {
      message.constructorInput = String(object.constructorInput);
    } else {
      message.constructorInput = "";
    }
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = String(object.chainID);
    } else {
      message.chainID = "";
    }
    return message;
  },

  toJSON(message: MsgUploadNewSmartContractTemp): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.abi !== undefined && (obj.abi = message.abi);
    message.bytecode !== undefined && (obj.bytecode = message.bytecode);
    message.constructorInput !== undefined &&
      (obj.constructorInput = message.constructorInput);
    message.chainID !== undefined && (obj.chainID = message.chainID);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgUploadNewSmartContractTemp>
  ): MsgUploadNewSmartContractTemp {
    const message = {
      ...baseMsgUploadNewSmartContractTemp,
    } as MsgUploadNewSmartContractTemp;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.abi !== undefined && object.abi !== null) {
      message.abi = object.abi;
    } else {
      message.abi = "";
    }
    if (object.bytecode !== undefined && object.bytecode !== null) {
      message.bytecode = object.bytecode;
    } else {
      message.bytecode = "";
    }
    if (
      object.constructorInput !== undefined &&
      object.constructorInput !== null
    ) {
      message.constructorInput = object.constructorInput;
    } else {
      message.constructorInput = "";
    }
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = object.chainID;
    } else {
      message.chainID = "";
    }
    return message;
  },
};

const baseMsgUploadNewSmartContractTempResponse: object = {};

export const MsgUploadNewSmartContractTempResponse = {
  encode(
    _: MsgUploadNewSmartContractTempResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgUploadNewSmartContractTempResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgUploadNewSmartContractTempResponse,
    } as MsgUploadNewSmartContractTempResponse;
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

  fromJSON(_: any): MsgUploadNewSmartContractTempResponse {
    const message = {
      ...baseMsgUploadNewSmartContractTempResponse,
    } as MsgUploadNewSmartContractTempResponse;
    return message;
  },

  toJSON(_: MsgUploadNewSmartContractTempResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgUploadNewSmartContractTempResponse>
  ): MsgUploadNewSmartContractTempResponse {
    const message = {
      ...baseMsgUploadNewSmartContractTempResponse,
    } as MsgUploadNewSmartContractTempResponse;
    return message;
  },
};

/** Msg defines the Msg service. */
export interface Msg {
  SubmitNewJob(request: MsgSubmitNewJob): Promise<MsgSubmitNewJobResponse>;
  /** this line is used by starport scaffolding # proto/tx/rpc */
  UploadNewSmartContractTemp(
    request: MsgUploadNewSmartContractTemp
  ): Promise<MsgUploadNewSmartContractTempResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  SubmitNewJob(request: MsgSubmitNewJob): Promise<MsgSubmitNewJobResponse> {
    const data = MsgSubmitNewJob.encode(request).finish();
    const promise = this.rpc.request(
      "palomachain.paloma.evm.Msg",
      "SubmitNewJob",
      data
    );
    return promise.then((data) =>
      MsgSubmitNewJobResponse.decode(new Reader(data))
    );
  }

  UploadNewSmartContractTemp(
    request: MsgUploadNewSmartContractTemp
  ): Promise<MsgUploadNewSmartContractTempResponse> {
    const data = MsgUploadNewSmartContractTemp.encode(request).finish();
    const promise = this.rpc.request(
      "palomachain.paloma.evm.Msg",
      "UploadNewSmartContractTemp",
      data
    );
    return promise.then((data) =>
      MsgUploadNewSmartContractTempResponse.decode(new Reader(data))
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
