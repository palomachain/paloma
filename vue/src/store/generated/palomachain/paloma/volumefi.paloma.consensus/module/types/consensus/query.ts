/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { Params } from "../consensus/params";
import { Any } from "../google/protobuf/any";

export const protobufPackage = "volumefi.paloma.consensus";

/** QueryParamsRequest is request type for the Query/Params RPC method. */
export interface QueryParamsRequest {}

/** QueryParamsResponse is response type for the Query/Params RPC method. */
export interface QueryParamsResponse {
  /** params holds all the parameters of this module. */
  params: Params | undefined;
}

export interface QueryQueuedMessagesForSigningRequest {
  valAddress: Uint8Array;
  queueTypeName: string;
}

export interface QueryQueuedMessagesForSigningResponse {
  messageToSign: MessageToSign[];
}

export interface MessageToSign {
  nonce: Uint8Array;
  id: number;
  bytesToSign: Uint8Array;
  msg: Any | undefined;
}

export interface MessageApprovedSignData {
  valAddress: Uint8Array;
  signature: Uint8Array;
}

export interface MessageApproved {
  nonce: Uint8Array;
  id: number;
  msg: Any | undefined;
  signData: MessageApprovedSignData[];
}

export interface QueryConsensusReachedRequest {
  queueTypeName: string;
}

export interface QueryConsensusReachedResponse {
  messages: MessageApproved[];
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

const baseQueryQueuedMessagesForSigningRequest: object = { queueTypeName: "" };

export const QueryQueuedMessagesForSigningRequest = {
  encode(
    message: QueryQueuedMessagesForSigningRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.valAddress.length !== 0) {
      writer.uint32(10).bytes(message.valAddress);
    }
    if (message.queueTypeName !== "") {
      writer.uint32(18).string(message.queueTypeName);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryQueuedMessagesForSigningRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryQueuedMessagesForSigningRequest,
    } as QueryQueuedMessagesForSigningRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.valAddress = reader.bytes();
          break;
        case 2:
          message.queueTypeName = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryQueuedMessagesForSigningRequest {
    const message = {
      ...baseQueryQueuedMessagesForSigningRequest,
    } as QueryQueuedMessagesForSigningRequest;
    if (object.valAddress !== undefined && object.valAddress !== null) {
      message.valAddress = bytesFromBase64(object.valAddress);
    }
    if (object.queueTypeName !== undefined && object.queueTypeName !== null) {
      message.queueTypeName = String(object.queueTypeName);
    } else {
      message.queueTypeName = "";
    }
    return message;
  },

  toJSON(message: QueryQueuedMessagesForSigningRequest): unknown {
    const obj: any = {};
    message.valAddress !== undefined &&
      (obj.valAddress = base64FromBytes(
        message.valAddress !== undefined ? message.valAddress : new Uint8Array()
      ));
    message.queueTypeName !== undefined &&
      (obj.queueTypeName = message.queueTypeName);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryQueuedMessagesForSigningRequest>
  ): QueryQueuedMessagesForSigningRequest {
    const message = {
      ...baseQueryQueuedMessagesForSigningRequest,
    } as QueryQueuedMessagesForSigningRequest;
    if (object.valAddress !== undefined && object.valAddress !== null) {
      message.valAddress = object.valAddress;
    } else {
      message.valAddress = new Uint8Array();
    }
    if (object.queueTypeName !== undefined && object.queueTypeName !== null) {
      message.queueTypeName = object.queueTypeName;
    } else {
      message.queueTypeName = "";
    }
    return message;
  },
};

const baseQueryQueuedMessagesForSigningResponse: object = {};

export const QueryQueuedMessagesForSigningResponse = {
  encode(
    message: QueryQueuedMessagesForSigningResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.messageToSign) {
      MessageToSign.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryQueuedMessagesForSigningResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryQueuedMessagesForSigningResponse,
    } as QueryQueuedMessagesForSigningResponse;
    message.messageToSign = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.messageToSign.push(
            MessageToSign.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryQueuedMessagesForSigningResponse {
    const message = {
      ...baseQueryQueuedMessagesForSigningResponse,
    } as QueryQueuedMessagesForSigningResponse;
    message.messageToSign = [];
    if (object.messageToSign !== undefined && object.messageToSign !== null) {
      for (const e of object.messageToSign) {
        message.messageToSign.push(MessageToSign.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: QueryQueuedMessagesForSigningResponse): unknown {
    const obj: any = {};
    if (message.messageToSign) {
      obj.messageToSign = message.messageToSign.map((e) =>
        e ? MessageToSign.toJSON(e) : undefined
      );
    } else {
      obj.messageToSign = [];
    }
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryQueuedMessagesForSigningResponse>
  ): QueryQueuedMessagesForSigningResponse {
    const message = {
      ...baseQueryQueuedMessagesForSigningResponse,
    } as QueryQueuedMessagesForSigningResponse;
    message.messageToSign = [];
    if (object.messageToSign !== undefined && object.messageToSign !== null) {
      for (const e of object.messageToSign) {
        message.messageToSign.push(MessageToSign.fromPartial(e));
      }
    }
    return message;
  },
};

const baseMessageToSign: object = { id: 0 };

export const MessageToSign = {
  encode(message: MessageToSign, writer: Writer = Writer.create()): Writer {
    if (message.nonce.length !== 0) {
      writer.uint32(10).bytes(message.nonce);
    }
    if (message.id !== 0) {
      writer.uint32(16).uint64(message.id);
    }
    if (message.bytesToSign.length !== 0) {
      writer.uint32(26).bytes(message.bytesToSign);
    }
    if (message.msg !== undefined) {
      Any.encode(message.msg, writer.uint32(34).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MessageToSign {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMessageToSign } as MessageToSign;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.nonce = reader.bytes();
          break;
        case 2:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        case 3:
          message.bytesToSign = reader.bytes();
          break;
        case 4:
          message.msg = Any.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MessageToSign {
    const message = { ...baseMessageToSign } as MessageToSign;
    if (object.nonce !== undefined && object.nonce !== null) {
      message.nonce = bytesFromBase64(object.nonce);
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    if (object.bytesToSign !== undefined && object.bytesToSign !== null) {
      message.bytesToSign = bytesFromBase64(object.bytesToSign);
    }
    if (object.msg !== undefined && object.msg !== null) {
      message.msg = Any.fromJSON(object.msg);
    } else {
      message.msg = undefined;
    }
    return message;
  },

  toJSON(message: MessageToSign): unknown {
    const obj: any = {};
    message.nonce !== undefined &&
      (obj.nonce = base64FromBytes(
        message.nonce !== undefined ? message.nonce : new Uint8Array()
      ));
    message.id !== undefined && (obj.id = message.id);
    message.bytesToSign !== undefined &&
      (obj.bytesToSign = base64FromBytes(
        message.bytesToSign !== undefined
          ? message.bytesToSign
          : new Uint8Array()
      ));
    message.msg !== undefined &&
      (obj.msg = message.msg ? Any.toJSON(message.msg) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<MessageToSign>): MessageToSign {
    const message = { ...baseMessageToSign } as MessageToSign;
    if (object.nonce !== undefined && object.nonce !== null) {
      message.nonce = object.nonce;
    } else {
      message.nonce = new Uint8Array();
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    if (object.bytesToSign !== undefined && object.bytesToSign !== null) {
      message.bytesToSign = object.bytesToSign;
    } else {
      message.bytesToSign = new Uint8Array();
    }
    if (object.msg !== undefined && object.msg !== null) {
      message.msg = Any.fromPartial(object.msg);
    } else {
      message.msg = undefined;
    }
    return message;
  },
};

const baseMessageApprovedSignData: object = {};

export const MessageApprovedSignData = {
  encode(
    message: MessageApprovedSignData,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.valAddress.length !== 0) {
      writer.uint32(10).bytes(message.valAddress);
    }
    if (message.signature.length !== 0) {
      writer.uint32(18).bytes(message.signature);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MessageApprovedSignData {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMessageApprovedSignData,
    } as MessageApprovedSignData;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.valAddress = reader.bytes();
          break;
        case 2:
          message.signature = reader.bytes();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MessageApprovedSignData {
    const message = {
      ...baseMessageApprovedSignData,
    } as MessageApprovedSignData;
    if (object.valAddress !== undefined && object.valAddress !== null) {
      message.valAddress = bytesFromBase64(object.valAddress);
    }
    if (object.signature !== undefined && object.signature !== null) {
      message.signature = bytesFromBase64(object.signature);
    }
    return message;
  },

  toJSON(message: MessageApprovedSignData): unknown {
    const obj: any = {};
    message.valAddress !== undefined &&
      (obj.valAddress = base64FromBytes(
        message.valAddress !== undefined ? message.valAddress : new Uint8Array()
      ));
    message.signature !== undefined &&
      (obj.signature = base64FromBytes(
        message.signature !== undefined ? message.signature : new Uint8Array()
      ));
    return obj;
  },

  fromPartial(
    object: DeepPartial<MessageApprovedSignData>
  ): MessageApprovedSignData {
    const message = {
      ...baseMessageApprovedSignData,
    } as MessageApprovedSignData;
    if (object.valAddress !== undefined && object.valAddress !== null) {
      message.valAddress = object.valAddress;
    } else {
      message.valAddress = new Uint8Array();
    }
    if (object.signature !== undefined && object.signature !== null) {
      message.signature = object.signature;
    } else {
      message.signature = new Uint8Array();
    }
    return message;
  },
};

const baseMessageApproved: object = { id: 0 };

export const MessageApproved = {
  encode(message: MessageApproved, writer: Writer = Writer.create()): Writer {
    if (message.nonce.length !== 0) {
      writer.uint32(10).bytes(message.nonce);
    }
    if (message.id !== 0) {
      writer.uint32(16).uint64(message.id);
    }
    if (message.msg !== undefined) {
      Any.encode(message.msg, writer.uint32(26).fork()).ldelim();
    }
    for (const v of message.signData) {
      MessageApprovedSignData.encode(v!, writer.uint32(34).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MessageApproved {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMessageApproved } as MessageApproved;
    message.signData = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.nonce = reader.bytes();
          break;
        case 2:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        case 3:
          message.msg = Any.decode(reader, reader.uint32());
          break;
        case 4:
          message.signData.push(
            MessageApprovedSignData.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MessageApproved {
    const message = { ...baseMessageApproved } as MessageApproved;
    message.signData = [];
    if (object.nonce !== undefined && object.nonce !== null) {
      message.nonce = bytesFromBase64(object.nonce);
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    if (object.msg !== undefined && object.msg !== null) {
      message.msg = Any.fromJSON(object.msg);
    } else {
      message.msg = undefined;
    }
    if (object.signData !== undefined && object.signData !== null) {
      for (const e of object.signData) {
        message.signData.push(MessageApprovedSignData.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: MessageApproved): unknown {
    const obj: any = {};
    message.nonce !== undefined &&
      (obj.nonce = base64FromBytes(
        message.nonce !== undefined ? message.nonce : new Uint8Array()
      ));
    message.id !== undefined && (obj.id = message.id);
    message.msg !== undefined &&
      (obj.msg = message.msg ? Any.toJSON(message.msg) : undefined);
    if (message.signData) {
      obj.signData = message.signData.map((e) =>
        e ? MessageApprovedSignData.toJSON(e) : undefined
      );
    } else {
      obj.signData = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<MessageApproved>): MessageApproved {
    const message = { ...baseMessageApproved } as MessageApproved;
    message.signData = [];
    if (object.nonce !== undefined && object.nonce !== null) {
      message.nonce = object.nonce;
    } else {
      message.nonce = new Uint8Array();
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    if (object.msg !== undefined && object.msg !== null) {
      message.msg = Any.fromPartial(object.msg);
    } else {
      message.msg = undefined;
    }
    if (object.signData !== undefined && object.signData !== null) {
      for (const e of object.signData) {
        message.signData.push(MessageApprovedSignData.fromPartial(e));
      }
    }
    return message;
  },
};

const baseQueryConsensusReachedRequest: object = { queueTypeName: "" };

export const QueryConsensusReachedRequest = {
  encode(
    message: QueryConsensusReachedRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.queueTypeName !== "") {
      writer.uint32(10).string(message.queueTypeName);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryConsensusReachedRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryConsensusReachedRequest,
    } as QueryConsensusReachedRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.queueTypeName = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryConsensusReachedRequest {
    const message = {
      ...baseQueryConsensusReachedRequest,
    } as QueryConsensusReachedRequest;
    if (object.queueTypeName !== undefined && object.queueTypeName !== null) {
      message.queueTypeName = String(object.queueTypeName);
    } else {
      message.queueTypeName = "";
    }
    return message;
  },

  toJSON(message: QueryConsensusReachedRequest): unknown {
    const obj: any = {};
    message.queueTypeName !== undefined &&
      (obj.queueTypeName = message.queueTypeName);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryConsensusReachedRequest>
  ): QueryConsensusReachedRequest {
    const message = {
      ...baseQueryConsensusReachedRequest,
    } as QueryConsensusReachedRequest;
    if (object.queueTypeName !== undefined && object.queueTypeName !== null) {
      message.queueTypeName = object.queueTypeName;
    } else {
      message.queueTypeName = "";
    }
    return message;
  },
};

const baseQueryConsensusReachedResponse: object = {};

export const QueryConsensusReachedResponse = {
  encode(
    message: QueryConsensusReachedResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.messages) {
      MessageApproved.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryConsensusReachedResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryConsensusReachedResponse,
    } as QueryConsensusReachedResponse;
    message.messages = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.messages.push(
            MessageApproved.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryConsensusReachedResponse {
    const message = {
      ...baseQueryConsensusReachedResponse,
    } as QueryConsensusReachedResponse;
    message.messages = [];
    if (object.messages !== undefined && object.messages !== null) {
      for (const e of object.messages) {
        message.messages.push(MessageApproved.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: QueryConsensusReachedResponse): unknown {
    const obj: any = {};
    if (message.messages) {
      obj.messages = message.messages.map((e) =>
        e ? MessageApproved.toJSON(e) : undefined
      );
    } else {
      obj.messages = [];
    }
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryConsensusReachedResponse>
  ): QueryConsensusReachedResponse {
    const message = {
      ...baseQueryConsensusReachedResponse,
    } as QueryConsensusReachedResponse;
    message.messages = [];
    if (object.messages !== undefined && object.messages !== null) {
      for (const e of object.messages) {
        message.messages.push(MessageApproved.fromPartial(e));
      }
    }
    return message;
  },
};

/** Query defines the gRPC querier service. */
export interface Query {
  /** Parameters queries the parameters of the module. */
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse>;
  /** Queries a list of QueuedMessagesForSigning items. */
  QueuedMessagesForSigning(
    request: QueryQueuedMessagesForSigningRequest
  ): Promise<QueryQueuedMessagesForSigningResponse>;
  /** Queries a list of ConsensusReached items. */
  ConsensusReached(
    request: QueryConsensusReachedRequest
  ): Promise<QueryConsensusReachedResponse>;
}

export class QueryClientImpl implements Query {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse> {
    const data = QueryParamsRequest.encode(request).finish();
    const promise = this.rpc.request(
      "volumefi.paloma.consensus.Query",
      "Params",
      data
    );
    return promise.then((data) => QueryParamsResponse.decode(new Reader(data)));
  }

  QueuedMessagesForSigning(
    request: QueryQueuedMessagesForSigningRequest
  ): Promise<QueryQueuedMessagesForSigningResponse> {
    const data = QueryQueuedMessagesForSigningRequest.encode(request).finish();
    const promise = this.rpc.request(
      "volumefi.paloma.consensus.Query",
      "QueuedMessagesForSigning",
      data
    );
    return promise.then((data) =>
      QueryQueuedMessagesForSigningResponse.decode(new Reader(data))
    );
  }

  ConsensusReached(
    request: QueryConsensusReachedRequest
  ): Promise<QueryConsensusReachedResponse> {
    const data = QueryConsensusReachedRequest.encode(request).finish();
    const promise = this.rpc.request(
      "volumefi.paloma.consensus.Query",
      "ConsensusReached",
      data
    );
    return promise.then((data) =>
      QueryConsensusReachedResponse.decode(new Reader(data))
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
