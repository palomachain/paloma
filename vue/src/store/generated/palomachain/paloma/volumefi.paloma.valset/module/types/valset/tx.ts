/* eslint-disable */
import { Reader, Writer } from "protobufjs/minimal";
import { ExternalChainInfo } from "../valset/snapshot";

export const protobufPackage = "volumefi.paloma.valset";

export interface MsgAddExternalChainInfoForValidator {
  creator: string;
  chainInfos: ExternalChainInfo[];
}

export interface MsgAddExternalChainInfoForValidatorResponse {}

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
      ExternalChainInfo.encode(v!, writer.uint32(18).fork()).ldelim();
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
            ExternalChainInfo.decode(reader, reader.uint32())
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
        message.chainInfos.push(ExternalChainInfo.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: MsgAddExternalChainInfoForValidator): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    if (message.chainInfos) {
      obj.chainInfos = message.chainInfos.map((e) =>
        e ? ExternalChainInfo.toJSON(e) : undefined
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
        message.chainInfos.push(ExternalChainInfo.fromPartial(e));
      }
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
