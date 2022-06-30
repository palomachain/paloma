/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { Params } from "../evm/params";
import { Valset } from "../evm/turnstone";

export const protobufPackage = "palomachain.paloma.evm";

/** QueryParamsRequest is request type for the Query/Params RPC method. */
export interface QueryParamsRequest {}

/** QueryParamsResponse is response type for the Query/Params RPC method. */
export interface QueryParamsResponse {
  /** params holds all the parameters of this module. */
  params: Params | undefined;
}

export interface QueryGetValsetByIDRequest {
  valsetID: number;
  chainID: string;
}

export interface QueryGetValsetByIDResponse {
  valset: Valset | undefined;
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

const baseQueryGetValsetByIDRequest: object = { valsetID: 0, chainID: "" };

export const QueryGetValsetByIDRequest = {
  encode(
    message: QueryGetValsetByIDRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.valsetID !== 0) {
      writer.uint32(8).uint64(message.valsetID);
    }
    if (message.chainID !== "") {
      writer.uint32(18).string(message.chainID);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetValsetByIDRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetValsetByIDRequest,
    } as QueryGetValsetByIDRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.valsetID = longToNumber(reader.uint64() as Long);
          break;
        case 2:
          message.chainID = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetValsetByIDRequest {
    const message = {
      ...baseQueryGetValsetByIDRequest,
    } as QueryGetValsetByIDRequest;
    if (object.valsetID !== undefined && object.valsetID !== null) {
      message.valsetID = Number(object.valsetID);
    } else {
      message.valsetID = 0;
    }
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = String(object.chainID);
    } else {
      message.chainID = "";
    }
    return message;
  },

  toJSON(message: QueryGetValsetByIDRequest): unknown {
    const obj: any = {};
    message.valsetID !== undefined && (obj.valsetID = message.valsetID);
    message.chainID !== undefined && (obj.chainID = message.chainID);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetValsetByIDRequest>
  ): QueryGetValsetByIDRequest {
    const message = {
      ...baseQueryGetValsetByIDRequest,
    } as QueryGetValsetByIDRequest;
    if (object.valsetID !== undefined && object.valsetID !== null) {
      message.valsetID = object.valsetID;
    } else {
      message.valsetID = 0;
    }
    if (object.chainID !== undefined && object.chainID !== null) {
      message.chainID = object.chainID;
    } else {
      message.chainID = "";
    }
    return message;
  },
};

const baseQueryGetValsetByIDResponse: object = {};

export const QueryGetValsetByIDResponse = {
  encode(
    message: QueryGetValsetByIDResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.valset !== undefined) {
      Valset.encode(message.valset, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetValsetByIDResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetValsetByIDResponse,
    } as QueryGetValsetByIDResponse;
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

  fromJSON(object: any): QueryGetValsetByIDResponse {
    const message = {
      ...baseQueryGetValsetByIDResponse,
    } as QueryGetValsetByIDResponse;
    if (object.valset !== undefined && object.valset !== null) {
      message.valset = Valset.fromJSON(object.valset);
    } else {
      message.valset = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetValsetByIDResponse): unknown {
    const obj: any = {};
    message.valset !== undefined &&
      (obj.valset = message.valset ? Valset.toJSON(message.valset) : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetValsetByIDResponse>
  ): QueryGetValsetByIDResponse {
    const message = {
      ...baseQueryGetValsetByIDResponse,
    } as QueryGetValsetByIDResponse;
    if (object.valset !== undefined && object.valset !== null) {
      message.valset = Valset.fromPartial(object.valset);
    } else {
      message.valset = undefined;
    }
    return message;
  },
};

/** Query defines the gRPC querier service. */
export interface Query {
  /** Parameters queries the parameters of the module. */
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse>;
  /** Queries a list of GetValsetByID items. */
  GetValsetByID(
    request: QueryGetValsetByIDRequest
  ): Promise<QueryGetValsetByIDResponse>;
}

export class QueryClientImpl implements Query {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse> {
    const data = QueryParamsRequest.encode(request).finish();
    const promise = this.rpc.request(
      "palomachain.paloma.evm.Query",
      "Params",
      data
    );
    return promise.then((data) => QueryParamsResponse.decode(new Reader(data)));
  }

  GetValsetByID(
    request: QueryGetValsetByIDRequest
  ): Promise<QueryGetValsetByIDResponse> {
    const data = QueryGetValsetByIDRequest.encode(request).finish();
    const promise = this.rpc.request(
      "palomachain.paloma.evm.Query",
      "GetValsetByID",
      data
    );
    return promise.then((data) =>
      QueryGetValsetByIDResponse.decode(new Reader(data))
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
