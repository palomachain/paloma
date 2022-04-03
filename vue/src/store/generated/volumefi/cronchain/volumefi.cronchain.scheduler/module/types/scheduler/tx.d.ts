import { Reader, Writer } from "protobufjs/minimal";
export declare const protobufPackage = "volumefi.cronchain.scheduler";
export interface MsgSigningQueueMessage {
    creator: string;
    id: string;
}
export interface MsgSigningQueueMessageResponse {
}
export declare const MsgSigningQueueMessage: {
    encode(message: MsgSigningQueueMessage, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgSigningQueueMessage;
    fromJSON(object: any): MsgSigningQueueMessage;
    toJSON(message: MsgSigningQueueMessage): unknown;
    fromPartial(object: DeepPartial<MsgSigningQueueMessage>): MsgSigningQueueMessage;
};
export declare const MsgSigningQueueMessageResponse: {
    encode(_: MsgSigningQueueMessageResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgSigningQueueMessageResponse;
    fromJSON(_: any): MsgSigningQueueMessageResponse;
    toJSON(_: MsgSigningQueueMessageResponse): unknown;
    fromPartial(_: DeepPartial<MsgSigningQueueMessageResponse>): MsgSigningQueueMessageResponse;
};
/** Msg defines the Msg service. */
export interface Msg {
    /** this line is used by starport scaffolding # proto/tx/rpc */
    SigningQueueMessage(request: MsgSigningQueueMessage): Promise<MsgSigningQueueMessageResponse>;
}
export declare class MsgClientImpl implements Msg {
    private readonly rpc;
    constructor(rpc: Rpc);
    SigningQueueMessage(request: MsgSigningQueueMessage): Promise<MsgSigningQueueMessageResponse>;
}
interface Rpc {
    request(service: string, method: string, data: Uint8Array): Promise<Uint8Array>;
}
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
