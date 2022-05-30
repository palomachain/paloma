/* eslint-disable */
import { Timestamp } from "../google/protobuf/timestamp";
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "volumefi.paloma.valset";

export enum ValidatorState {
  NONE = 0,
  ACTIVE = 1,
  JAILED = 2,
  UNRECOGNIZED = -1,
}

export function validatorStateFromJSON(object: any): ValidatorState {
  switch (object) {
    case 0:
    case "NONE":
      return ValidatorState.NONE;
    case 1:
    case "ACTIVE":
      return ValidatorState.ACTIVE;
    case 2:
    case "JAILED":
      return ValidatorState.JAILED;
    case -1:
    case "UNRECOGNIZED":
    default:
      return ValidatorState.UNRECOGNIZED;
  }
}

export function validatorStateToJSON(object: ValidatorState): string {
  switch (object) {
    case ValidatorState.NONE:
      return "NONE";
    case ValidatorState.ACTIVE:
      return "ACTIVE";
    case ValidatorState.JAILED:
      return "JAILED";
    default:
      return "UNKNOWN";
  }
}

export interface Validator {
  shareCount: Uint8Array;
  /** TODO: make this ed25519 pub key instead of bytes */
  pubKey: Uint8Array;
  state: ValidatorState;
  externalChainInfos: ExternalChainInfo[];
  address: Uint8Array;
  signerAddress: Uint8Array;
}

export interface Snapshot {
  validators: Validator[];
  height: number;
  totalShares: Uint8Array;
  createdAt: Date | undefined;
}

export interface ExternalChainInfo {
  ID: number;
  chainID: string;
  address: string;
}

const baseValidator: object = { state: 0 };

export const Validator = {
  encode(message: Validator, writer: Writer = Writer.create()): Writer {
    if (message.shareCount.length !== 0) {
      writer.uint32(10).bytes(message.shareCount);
    }
    if (message.pubKey.length !== 0) {
      writer.uint32(18).bytes(message.pubKey);
    }
    if (message.state !== 0) {
      writer.uint32(24).int32(message.state);
    }
    for (const v of message.externalChainInfos) {
      ExternalChainInfo.encode(v!, writer.uint32(34).fork()).ldelim();
    }
    if (message.address.length !== 0) {
      writer.uint32(42).bytes(message.address);
    }
    if (message.signerAddress.length !== 0) {
      writer.uint32(50).bytes(message.signerAddress);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Validator {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseValidator } as Validator;
    message.externalChainInfos = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.shareCount = reader.bytes();
          break;
        case 2:
          message.pubKey = reader.bytes();
          break;
        case 3:
          message.state = reader.int32() as any;
          break;
        case 4:
          message.externalChainInfos.push(
            ExternalChainInfo.decode(reader, reader.uint32())
          );
          break;
        case 5:
          message.address = reader.bytes();
          break;
        case 6:
          message.signerAddress = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Validator {
    const message = { ...baseValidator } as Validator;
    message.externalChainInfos = [];
    if (object.shareCount !== undefined && object.shareCount !== null) {
      message.shareCount = bytesFromBase64(object.shareCount);
    }
    if (object.pubKey !== undefined && object.pubKey !== null) {
      message.pubKey = bytesFromBase64(object.pubKey);
    }
    if (object.state !== undefined && object.state !== null) {
      message.state = validatorStateFromJSON(object.state);
    } else {
      message.state = 0;
    }
    if (
      object.externalChainInfos !== undefined &&
      object.externalChainInfos !== null
    ) {
      for (const e of object.externalChainInfos) {
        message.externalChainInfos.push(ExternalChainInfo.fromJSON(e));
      }
    }
    if (object.address !== undefined && object.address !== null) {
      message.address = bytesFromBase64(object.address);
    }
    if (object.signerAddress !== undefined && object.signerAddress !== null) {
      message.signerAddress = bytesFromBase64(object.signerAddress);
    }
    return message;
  },

  toJSON(message: Validator): unknown {
    const obj: any = {};
    message.shareCount !== undefined &&
      (obj.shareCount = base64FromBytes(
        message.shareCount !== undefined ? message.shareCount : new Uint8Array()
      ));
    message.pubKey !== undefined &&
      (obj.pubKey = base64FromBytes(
        message.pubKey !== undefined ? message.pubKey : new Uint8Array()
      ));
    message.state !== undefined &&
      (obj.state = validatorStateToJSON(message.state));
    if (message.externalChainInfos) {
      obj.externalChainInfos = message.externalChainInfos.map((e) =>
        e ? ExternalChainInfo.toJSON(e) : undefined
      );
    } else {
      obj.externalChainInfos = [];
    }
    message.address !== undefined &&
      (obj.address = base64FromBytes(
        message.address !== undefined ? message.address : new Uint8Array()
      ));
    message.signerAddress !== undefined &&
      (obj.signerAddress = base64FromBytes(
        message.signerAddress !== undefined
          ? message.signerAddress
          : new Uint8Array()
      ));
    return obj;
  },

  fromPartial(object: DeepPartial<Validator>): Validator {
    const message = { ...baseValidator } as Validator;
    message.externalChainInfos = [];
    if (object.shareCount !== undefined && object.shareCount !== null) {
      message.shareCount = object.shareCount;
    } else {
      message.shareCount = new Uint8Array();
    }
    if (object.pubKey !== undefined && object.pubKey !== null) {
      message.pubKey = object.pubKey;
    } else {
      message.pubKey = new Uint8Array();
    }
    if (object.state !== undefined && object.state !== null) {
      message.state = object.state;
    } else {
      message.state = 0;
    }
    if (
      object.externalChainInfos !== undefined &&
      object.externalChainInfos !== null
    ) {
      for (const e of object.externalChainInfos) {
        message.externalChainInfos.push(ExternalChainInfo.fromPartial(e));
      }
    }
    if (object.address !== undefined && object.address !== null) {
      message.address = object.address;
    } else {
      message.address = new Uint8Array();
    }
    if (object.signerAddress !== undefined && object.signerAddress !== null) {
      message.signerAddress = object.signerAddress;
    } else {
      message.signerAddress = new Uint8Array();
    }
    return message;
  },
};

const baseSnapshot: object = { height: 0 };

export const Snapshot = {
  encode(message: Snapshot, writer: Writer = Writer.create()): Writer {
    for (const v of message.validators) {
      Validator.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.height !== 0) {
      writer.uint32(16).int64(message.height);
    }
    if (message.totalShares.length !== 0) {
      writer.uint32(26).bytes(message.totalShares);
    }
    if (message.createdAt !== undefined) {
      Timestamp.encode(
        toTimestamp(message.createdAt),
        writer.uint32(34).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Snapshot {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseSnapshot } as Snapshot;
    message.validators = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.validators.push(Validator.decode(reader, reader.uint32()));
          break;
        case 2:
          message.height = longToNumber(reader.int64() as Long);
          break;
        case 3:
          message.totalShares = reader.bytes();
          break;
        case 4:
          message.createdAt = fromTimestamp(
            Timestamp.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Snapshot {
    const message = { ...baseSnapshot } as Snapshot;
    message.validators = [];
    if (object.validators !== undefined && object.validators !== null) {
      for (const e of object.validators) {
        message.validators.push(Validator.fromJSON(e));
      }
    }
    if (object.height !== undefined && object.height !== null) {
      message.height = Number(object.height);
    } else {
      message.height = 0;
    }
    if (object.totalShares !== undefined && object.totalShares !== null) {
      message.totalShares = bytesFromBase64(object.totalShares);
    }
    if (object.createdAt !== undefined && object.createdAt !== null) {
      message.createdAt = fromJsonTimestamp(object.createdAt);
    } else {
      message.createdAt = undefined;
    }
    return message;
  },

  toJSON(message: Snapshot): unknown {
    const obj: any = {};
    if (message.validators) {
      obj.validators = message.validators.map((e) =>
        e ? Validator.toJSON(e) : undefined
      );
    } else {
      obj.validators = [];
    }
    message.height !== undefined && (obj.height = message.height);
    message.totalShares !== undefined &&
      (obj.totalShares = base64FromBytes(
        message.totalShares !== undefined
          ? message.totalShares
          : new Uint8Array()
      ));
    message.createdAt !== undefined &&
      (obj.createdAt =
        message.createdAt !== undefined
          ? message.createdAt.toISOString()
          : null);
    return obj;
  },

  fromPartial(object: DeepPartial<Snapshot>): Snapshot {
    const message = { ...baseSnapshot } as Snapshot;
    message.validators = [];
    if (object.validators !== undefined && object.validators !== null) {
      for (const e of object.validators) {
        message.validators.push(Validator.fromPartial(e));
      }
    }
    if (object.height !== undefined && object.height !== null) {
      message.height = object.height;
    } else {
      message.height = 0;
    }
    if (object.totalShares !== undefined && object.totalShares !== null) {
      message.totalShares = object.totalShares;
    } else {
      message.totalShares = new Uint8Array();
    }
    if (object.createdAt !== undefined && object.createdAt !== null) {
      message.createdAt = object.createdAt;
    } else {
      message.createdAt = undefined;
    }
    return message;
  },
};

const baseExternalChainInfo: object = { ID: 0, chainID: "", address: "" };

export const ExternalChainInfo = {
  encode(message: ExternalChainInfo, writer: Writer = Writer.create()): Writer {
    if (message.ID !== 0) {
      writer.uint32(8).uint64(message.ID);
    }
    if (message.chainID !== "") {
      writer.uint32(18).string(message.chainID);
    }
    if (message.address !== "") {
      writer.uint32(26).string(message.address);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): ExternalChainInfo {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseExternalChainInfo } as ExternalChainInfo;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.ID = longToNumber(reader.uint64() as Long);
          break;
        case 2:
          message.chainID = reader.string();
          break;
        case 3:
          message.address = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): ExternalChainInfo {
    const message = { ...baseExternalChainInfo } as ExternalChainInfo;
    if (object.ID !== undefined && object.ID !== null) {
      message.ID = Number(object.ID);
    } else {
      message.ID = 0;
    }
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

  toJSON(message: ExternalChainInfo): unknown {
    const obj: any = {};
    message.ID !== undefined && (obj.ID = message.ID);
    message.chainID !== undefined && (obj.chainID = message.chainID);
    message.address !== undefined && (obj.address = message.address);
    return obj;
  },

  fromPartial(object: DeepPartial<ExternalChainInfo>): ExternalChainInfo {
    const message = { ...baseExternalChainInfo } as ExternalChainInfo;
    if (object.ID !== undefined && object.ID !== null) {
      message.ID = object.ID;
    } else {
      message.ID = 0;
    }
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

function toTimestamp(date: Date): Timestamp {
  const seconds = date.getTime() / 1_000;
  const nanos = (date.getTime() % 1_000) * 1_000_000;
  return { seconds, nanos };
}

function fromTimestamp(t: Timestamp): Date {
  let millis = t.seconds * 1_000;
  millis += t.nanos / 1_000_000;
  return new Date(millis);
}

function fromJsonTimestamp(o: any): Date {
  if (o instanceof Date) {
    return o;
  } else if (typeof o === "string") {
    return new Date(o);
  } else {
    return fromTimestamp(Timestamp.fromJSON(o));
  }
}

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
