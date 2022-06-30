/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { Params } from "../valset/params";
import { ExternalChainInfo, Snapshot } from "../valset/snapshot";

export const protobufPackage = "volumefi.paloma.valset";

/** QueryParamsRequest is request type for the Query/Params RPC method. */
export interface QueryParamsRequest {}

/** QueryParamsResponse is response type for the Query/Params RPC method. */
export interface QueryParamsResponse {
  /** params holds all the parameters of this module. */
  params: Params | undefined;
}

export interface QueryValidatorInfoRequest {
  valAddr: string;
}

export interface QueryValidatorInfoResponse {
  chainInfos: ExternalChainInfo[];
}

export interface QueryGetSnapshotByIDRequest {
  snapshotId: number;
}

export interface QueryGetSnapshotByIDResponse {
  snapshot: Snapshot | undefined;
}

const baseQueryParamsRequest: object = {};

export const QueryParamsRequest = {
  encode(_: QueryParamsRequest, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryParamsRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
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

  fromJSON(_: any): QueryParamsRequest {
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    return message;
  },

  toJSON(_: QueryParamsRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<QueryParamsRequest>): QueryParamsRequest {
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    return message;
  },
};

const baseQueryParamsResponse: object = {};

export const QueryParamsResponse = {
  encode(
    message: QueryParamsResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.params !== undefined) {
      Params.encode(message.params, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryParamsResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.params = Params.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryParamsResponse {
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromJSON(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },

  toJSON(message: QueryParamsResponse): unknown {
    const obj: any = {};
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryParamsResponse>): QueryParamsResponse {
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromPartial(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },
};

const baseQueryValidatorInfoRequest: object = { valAddr: "" };

export const QueryValidatorInfoRequest = {
  encode(
    message: QueryValidatorInfoRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.valAddr !== "") {
      writer.uint32(10).string(message.valAddr);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryValidatorInfoRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryValidatorInfoRequest,
    } as QueryValidatorInfoRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.valAddr = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryValidatorInfoRequest {
    const message = {
      ...baseQueryValidatorInfoRequest,
    } as QueryValidatorInfoRequest;
    if (object.valAddr !== undefined && object.valAddr !== null) {
      message.valAddr = String(object.valAddr);
    } else {
      message.valAddr = "";
    }
    return message;
  },

  toJSON(message: QueryValidatorInfoRequest): unknown {
    const obj: any = {};
    message.valAddr !== undefined && (obj.valAddr = message.valAddr);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryValidatorInfoRequest>
  ): QueryValidatorInfoRequest {
    const message = {
      ...baseQueryValidatorInfoRequest,
    } as QueryValidatorInfoRequest;
    if (object.valAddr !== undefined && object.valAddr !== null) {
      message.valAddr = object.valAddr;
    } else {
      message.valAddr = "";
    }
    return message;
  },
};

const baseQueryValidatorInfoResponse: object = {};

export const QueryValidatorInfoResponse = {
  encode(
    message: QueryValidatorInfoResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.chainInfos) {
      ExternalChainInfo.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryValidatorInfoResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryValidatorInfoResponse,
    } as QueryValidatorInfoResponse;
    message.chainInfos = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
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

  fromJSON(object: any): QueryValidatorInfoResponse {
    const message = {
      ...baseQueryValidatorInfoResponse,
    } as QueryValidatorInfoResponse;
    message.chainInfos = [];
    if (object.chainInfos !== undefined && object.chainInfos !== null) {
      for (const e of object.chainInfos) {
        message.chainInfos.push(ExternalChainInfo.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: QueryValidatorInfoResponse): unknown {
    const obj: any = {};
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
    object: DeepPartial<QueryValidatorInfoResponse>
  ): QueryValidatorInfoResponse {
    const message = {
      ...baseQueryValidatorInfoResponse,
    } as QueryValidatorInfoResponse;
    message.chainInfos = [];
    if (object.chainInfos !== undefined && object.chainInfos !== null) {
      for (const e of object.chainInfos) {
        message.chainInfos.push(ExternalChainInfo.fromPartial(e));
      }
    }
    return message;
  },
};

const baseQueryGetSnapshotByIDRequest: object = { snapshotId: 0 };

export const QueryGetSnapshotByIDRequest = {
  encode(
    message: QueryGetSnapshotByIDRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.snapshotId !== 0) {
      writer.uint32(8).uint64(message.snapshotId);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetSnapshotByIDRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetSnapshotByIDRequest,
    } as QueryGetSnapshotByIDRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.snapshotId = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetSnapshotByIDRequest {
    const message = {
      ...baseQueryGetSnapshotByIDRequest,
    } as QueryGetSnapshotByIDRequest;
    if (object.snapshotId !== undefined && object.snapshotId !== null) {
      message.snapshotId = Number(object.snapshotId);
    } else {
      message.snapshotId = 0;
    }
    return message;
  },

  toJSON(message: QueryGetSnapshotByIDRequest): unknown {
    const obj: any = {};
    message.snapshotId !== undefined && (obj.snapshotId = message.snapshotId);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetSnapshotByIDRequest>
  ): QueryGetSnapshotByIDRequest {
    const message = {
      ...baseQueryGetSnapshotByIDRequest,
    } as QueryGetSnapshotByIDRequest;
    if (object.snapshotId !== undefined && object.snapshotId !== null) {
      message.snapshotId = object.snapshotId;
    } else {
      message.snapshotId = 0;
    }
    return message;
  },
};

const baseQueryGetSnapshotByIDResponse: object = {};

export const QueryGetSnapshotByIDResponse = {
  encode(
    message: QueryGetSnapshotByIDResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.snapshot !== undefined) {
      Snapshot.encode(message.snapshot, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetSnapshotByIDResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetSnapshotByIDResponse,
    } as QueryGetSnapshotByIDResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.snapshot = Snapshot.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetSnapshotByIDResponse {
    const message = {
      ...baseQueryGetSnapshotByIDResponse,
    } as QueryGetSnapshotByIDResponse;
    if (object.snapshot !== undefined && object.snapshot !== null) {
      message.snapshot = Snapshot.fromJSON(object.snapshot);
    } else {
      message.snapshot = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetSnapshotByIDResponse): unknown {
    const obj: any = {};
    message.snapshot !== undefined &&
      (obj.snapshot = message.snapshot
        ? Snapshot.toJSON(message.snapshot)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetSnapshotByIDResponse>
  ): QueryGetSnapshotByIDResponse {
    const message = {
      ...baseQueryGetSnapshotByIDResponse,
    } as QueryGetSnapshotByIDResponse;
    if (object.snapshot !== undefined && object.snapshot !== null) {
      message.snapshot = Snapshot.fromPartial(object.snapshot);
    } else {
      message.snapshot = undefined;
    }
    return message;
  },
};

/** Query defines the gRPC querier service. */
export interface Query {
  /** Parameters queries the parameters of the module. */
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse>;
  /** Queries a list of ValidatorInfo items. */
  ValidatorInfo(
    request: QueryValidatorInfoRequest
  ): Promise<QueryValidatorInfoResponse>;
  /** Queries a list of GetSnapshotByID items. */
  GetSnapshotByID(
    request: QueryGetSnapshotByIDRequest
  ): Promise<QueryGetSnapshotByIDResponse>;
}

export class QueryClientImpl implements Query {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse> {
    const data = QueryParamsRequest.encode(request).finish();
    const promise = this.rpc.request(
      "volumefi.paloma.valset.Query",
      "Params",
      data
    );
    return promise.then((data) => QueryParamsResponse.decode(new Reader(data)));
  }

  ValidatorInfo(
    request: QueryValidatorInfoRequest
  ): Promise<QueryValidatorInfoResponse> {
    const data = QueryValidatorInfoRequest.encode(request).finish();
    const promise = this.rpc.request(
      "volumefi.paloma.valset.Query",
      "ValidatorInfo",
      data
    );
    return promise.then((data) =>
      QueryValidatorInfoResponse.decode(new Reader(data))
    );
  }

  GetSnapshotByID(
    request: QueryGetSnapshotByIDRequest
  ): Promise<QueryGetSnapshotByIDResponse> {
    const data = QueryGetSnapshotByIDRequest.encode(request).finish();
    const promise = this.rpc.request(
      "volumefi.paloma.valset.Query",
      "GetSnapshotByID",
      data
    );
    return promise.then((data) =>
      QueryGetSnapshotByIDResponse.decode(new Reader(data))
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
