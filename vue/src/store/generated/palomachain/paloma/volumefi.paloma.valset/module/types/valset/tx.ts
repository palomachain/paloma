/* eslint-disable */
import { Reader, Writer } from "protobufjs/minimal";

export const protobufPackage = "volumefi.paloma.valset";

export interface MsgRegisterConductor {
  creator: string;
  /** TODO: make this a real pubkey type */
  pubKey: Uint8Array;
  valAddr: string;
  /**
   * caller should send us a signature of the byte representation of the pub
   * key they have sent us, so that we can validate that the pub key is
   * actually valid cryptographic key
   */
  signedPubKey: Uint8Array;
}

export interface MsgRegisterConductorResponse {}

export interface MsgAddExternalChainInfoForValidator {
  creator: string;
  chainInfos: MsgAddExternalChainInfoForValidator_ChainInfo[];
}

export interface MsgAddExternalChainInfoForValidator_ChainInfo {
  chainID: string;
  address: string;
}

export interface MsgAddExternalChainInfoForValidatorResponse {}

const baseMsgRegisterConductor: object = { creator: "", valAddr: "" };

export const MsgRegisterConductor = {
  encode(
    message: MsgRegisterConductor,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.pubKey.length !== 0) {
      writer.uint32(18).bytes(message.pubKey);
    }
    if (message.valAddr !== "") {
      writer.uint32(26).string(message.valAddr);
    }
    if (message.signedPubKey.length !== 0) {
      writer.uint32(34).bytes(message.signedPubKey);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgRegisterConductor {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgRegisterConductor } as MsgRegisterConductor;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.pubKey = reader.bytes();
          break;
        case 3:
          message.valAddr = reader.string();
          break;
        case 4:
          message.signedPubKey = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgRegisterConductor {
    const message = { ...baseMsgRegisterConductor } as MsgRegisterConductor;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.pubKey !== undefined && object.pubKey !== null) {
      message.pubKey = bytesFromBase64(object.pubKey);
    }
    if (object.valAddr !== undefined && object.valAddr !== null) {
      message.valAddr = String(object.valAddr);
    } else {
      message.valAddr = "";
    }
    if (object.signedPubKey !== undefined && object.signedPubKey !== null) {
      message.signedPubKey = bytesFromBase64(object.signedPubKey);
    }
    return message;
  },

  toJSON(message: MsgRegisterConductor): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.pubKey !== undefined &&
      (obj.pubKey = base64FromBytes(
        message.pubKey !== undefined ? message.pubKey : new Uint8Array()
      ));
    message.valAddr !== undefined && (obj.valAddr = message.valAddr);
    message.signedPubKey !== undefined &&
      (obj.signedPubKey = base64FromBytes(
        message.signedPubKey !== undefined
          ? message.signedPubKey
          : new Uint8Array()
      ));
    return obj;
  },

  fromPartial(object: DeepPartial<MsgRegisterConductor>): MsgRegisterConductor {
    const message = { ...baseMsgRegisterConductor } as MsgRegisterConductor;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.pubKey !== undefined && object.pubKey !== null) {
      message.pubKey = object.pubKey;
    } else {
      message.pubKey = new Uint8Array();
    }
    if (object.valAddr !== undefined && object.valAddr !== null) {
      message.valAddr = object.valAddr;
    } else {
      message.valAddr = "";
    }
    if (object.signedPubKey !== undefined && object.signedPubKey !== null) {
      message.signedPubKey = object.signedPubKey;
    } else {
      message.signedPubKey = new Uint8Array();
    }
    return message;
  },
};

const baseMsgRegisterConductorResponse: object = {};

export const MsgRegisterConductorResponse = {
  encode(
    _: MsgRegisterConductorResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgRegisterConductorResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgRegisterConductorResponse,
    } as MsgRegisterConductorResponse;
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

  fromJSON(_: any): MsgRegisterConductorResponse {
    const message = {
      ...baseMsgRegisterConductorResponse,
    } as MsgRegisterConductorResponse;
    return message;
  },

  toJSON(_: MsgRegisterConductorResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgRegisterConductorResponse>
  ): MsgRegisterConductorResponse {
    const message = {
      ...baseMsgRegisterConductorResponse,
    } as MsgRegisterConductorResponse;
    return message;
  },
};

const baseMsgAddExternalChainInfoForValidator: object = { creator: "" };

export const MsgAddExternalChainInfoForValidator = {
  encode(
    message: MsgAddExternalChainInfoForValidator,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    for (const v of message.chainInfos) {
      MsgAddExternalChainInfoForValidator_ChainInfo.encode(
        v!,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgAddExternalChainInfoForValidator {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgAddExternalChainInfoForValidator,
    } as MsgAddExternalChainInfoForValidator;
    message.chainInfos = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.chainInfos.push(
            MsgAddExternalChainInfoForValidator_ChainInfo.decode(
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

  fromJSON(object: any): MsgAddExternalChainInfoForValidator {
    const message = {
      ...baseMsgAddExternalChainInfoForValidator,
    } as MsgAddExternalChainInfoForValidator;
    message.chainInfos = [];
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.chainInfos !== undefined && object.chainInfos !== null) {
      for (const e of object.chainInfos) {
        message.chainInfos.push(
          MsgAddExternalChainInfoForValidator_ChainInfo.fromJSON(e)
        );
      }
    }
    return message;
  },

  toJSON(message: MsgAddExternalChainInfoForValidator): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    if (message.chainInfos) {
      obj.chainInfos = message.chainInfos.map((e) =>
        e ? MsgAddExternalChainInfoForValidator_ChainInfo.toJSON(e) : undefined
      );
    } else {
      obj.chainInfos = [];
    }
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgAddExternalChainInfoForValidator>
  ): MsgAddExternalChainInfoForValidator {
    const message = {
      ...baseMsgAddExternalChainInfoForValidator,
    } as MsgAddExternalChainInfoForValidator;
    message.chainInfos = [];
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.chainInfos !== undefined && object.chainInfos !== null) {
      for (const e of object.chainInfos) {
        message.chainInfos.push(
          MsgAddExternalChainInfoForValidator_ChainInfo.fromPartial(e)
        );
      }
    }
    return message;
  },
};

const baseMsgAddExternalChainInfoForValidator_ChainInfo: object = {
  chainID: "",
  address: "",
};

export const MsgAddExternalChainInfoForValidator_ChainInfo = {
  encode(
    message: MsgAddExternalChainInfoForValidator_ChainInfo,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.chainID !== "") {
      writer.uint32(10).string(message.chainID);
    }
    if (message.address !== "") {
      writer.uint32(18).string(message.address);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgAddExternalChainInfoForValidator_ChainInfo {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgAddExternalChainInfoForValidator_ChainInfo,
    } as MsgAddExternalChainInfoForValidator_ChainInfo;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.chainID = reader.string();
          break;
        case 2:
          message.address = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgAddExternalChainInfoForValidator_ChainInfo {
    const message = {
      ...baseMsgAddExternalChainInfoForValidator_ChainInfo,
    } as MsgAddExternalChainInfoForValidator_ChainInfo;
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = String(object.chainID);
    } else {
      message.chainID = "";
    }
    if (object.address !== undefined && object.address !== null) {
      message.address = String(object.address);
    } else {
      message.address = "";
    }
    return message;
  },

  toJSON(message: MsgAddExternalChainInfoForValidator_ChainInfo): unknown {
    const obj: any = {};
    message.chainID !== undefined && (obj.chainID = message.chainID);
    message.address !== undefined && (obj.address = message.address);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgAddExternalChainInfoForValidator_ChainInfo>
  ): MsgAddExternalChainInfoForValidator_ChainInfo {
    const message = {
      ...baseMsgAddExternalChainInfoForValidator_ChainInfo,
    } as MsgAddExternalChainInfoForValidator_ChainInfo;
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = object.chainID;
    } else {
      message.chainID = "";
    }
    if (object.address !== undefined && object.address !== null) {
      message.address = object.address;
    } else {
      message.address = "";
    }
    return message;
  },
};

const baseMsgAddExternalChainInfoForValidatorResponse: object = {};

export const MsgAddExternalChainInfoForValidatorResponse = {
  encode(
    _: MsgAddExternalChainInfoForValidatorResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgAddExternalChainInfoForValidatorResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgAddExternalChainInfoForValidatorResponse,
    } as MsgAddExternalChainInfoForValidatorResponse;
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

  fromJSON(_: any): MsgAddExternalChainInfoForValidatorResponse {
    const message = {
      ...baseMsgAddExternalChainInfoForValidatorResponse,
    } as MsgAddExternalChainInfoForValidatorResponse;
    return message;
  },

  toJSON(_: MsgAddExternalChainInfoForValidatorResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgAddExternalChainInfoForValidatorResponse>
  ): MsgAddExternalChainInfoForValidatorResponse {
    const message = {
      ...baseMsgAddExternalChainInfoForValidatorResponse,
    } as MsgAddExternalChainInfoForValidatorResponse;
    return message;
  },
};

/** Msg defines the Msg service. */
export interface Msg {
  RegisterConductor(
    request: MsgRegisterConductor
  ): Promise<MsgRegisterConductorResponse>;
  /** this line is used by starport scaffolding # proto/tx/rpc */
  AddExternalChainInfoForValidator(
    request: MsgAddExternalChainInfoForValidator
  ): Promise<MsgAddExternalChainInfoForValidatorResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  RegisterConductor(
    request: MsgRegisterConductor
  ): Promise<MsgRegisterConductorResponse> {
    const data = MsgRegisterConductor.encode(request).finish();
    const promise = this.rpc.request(
      "volumefi.paloma.valset.Msg",
      "RegisterConductor",
      data
    );
    return promise.then((data) =>
      MsgRegisterConductorResponse.decode(new Reader(data))
    );
  }

  AddExternalChainInfoForValidator(
    request: MsgAddExternalChainInfoForValidator
  ): Promise<MsgAddExternalChainInfoForValidatorResponse> {
    const data = MsgAddExternalChainInfoForValidator.encode(request).finish();
    const promise = this.rpc.request(
      "volumefi.paloma.valset.Msg",
      "AddExternalChainInfoForValidator",
      data
    );
    return promise.then((data) =>
      MsgAddExternalChainInfoForValidatorResponse.decode(new Reader(data))
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
