import { txClient, queryClient, MissingWalletError , registry} from './module'

import { QueuedSignedMessage } from "./module/types/consensus/consensus_queue"
import { BatchOfConsensusMessages } from "./module/types/consensus/consensus_queue"
import { Batch } from "./module/types/consensus/consensus_queue"
import { SignData } from "./module/types/consensus/consensus_queue"
import { ConsensusPacketData } from "./module/types/consensus/packet"
import { NoData } from "./module/types/consensus/packet"
import { Params } from "./module/types/consensus/params"
import { MessageToSign } from "./module/types/consensus/query"
import { ValidatorSignature } from "./module/types/consensus/query"
import { MessageWithSignatures } from "./module/types/consensus/query"
import { SignSmartContractExecute } from "./module/types/consensus/signing_messages"
import { MsgAddMessagesSignatures_MsgSignedMessage } from "./module/types/consensus/tx"


export { QueuedSignedMessage, BatchOfConsensusMessages, Batch, SignData, ConsensusPacketData, NoData, Params, MessageToSign, ValidatorSignature, MessageWithSignatures, SignSmartContractExecute, MsgAddMessagesSignatures_MsgSignedMessage };

async function initTxClient(vuexGetters) {
	return await txClient(vuexGetters['common/wallet/signer'], {
		addr: vuexGetters['common/env/apiTendermint']
	})
}

async function initQueryClient(vuexGetters) {
	return await queryClient({
		addr: vuexGetters['common/env/apiCosmos']
	})
}

function mergeResults(value, next_values) {
	for (let prop of Object.keys(next_values)) {
		if (Array.isArray(next_values[prop])) {
			value[prop]=[...value[prop], ...next_values[prop]]
		}else{
			value[prop]=next_values[prop]
		}
	}
	return value
}

function getStructure(template) {
	let structure = { fields: [] }
	for (const [key, value] of Object.entries(template)) {
		let field: any = {}
		field.name = key
		field.type = typeof value
		structure.fields.push(field)
	}
	return structure
}

const getDefaultState = () => {
	return {
				Params: {},
				QueuedMessagesForSigning: {},
				MessagesInQueue: {},
				GetAllQueueNames: {},
				
				_Structure: {
						QueuedSignedMessage: getStructure(QueuedSignedMessage.fromPartial({})),
						BatchOfConsensusMessages: getStructure(BatchOfConsensusMessages.fromPartial({})),
						Batch: getStructure(Batch.fromPartial({})),
						SignData: getStructure(SignData.fromPartial({})),
						ConsensusPacketData: getStructure(ConsensusPacketData.fromPartial({})),
						NoData: getStructure(NoData.fromPartial({})),
						Params: getStructure(Params.fromPartial({})),
						MessageToSign: getStructure(MessageToSign.fromPartial({})),
						ValidatorSignature: getStructure(ValidatorSignature.fromPartial({})),
						MessageWithSignatures: getStructure(MessageWithSignatures.fromPartial({})),
						SignSmartContractExecute: getStructure(SignSmartContractExecute.fromPartial({})),
						MsgAddMessagesSignatures_MsgSignedMessage: getStructure(MsgAddMessagesSignatures_MsgSignedMessage.fromPartial({})),
						
		},
		_Registry: registry,
		_Subscriptions: new Set(),
	}
}

// initial state
const state = getDefaultState()

export default {
	namespaced: true,
	state,
	mutations: {
		RESET_STATE(state) {
			Object.assign(state, getDefaultState())
		},
		QUERY(state, { query, key, value }) {
			state[query][JSON.stringify(key)] = value
		},
		SUBSCRIBE(state, subscription) {
			state._Subscriptions.add(JSON.stringify(subscription))
		},
		UNSUBSCRIBE(state, subscription) {
			state._Subscriptions.delete(JSON.stringify(subscription))
		}
	},
	getters: {
				getParams: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Params[JSON.stringify(params)] ?? {}
		},
				getQueuedMessagesForSigning: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.QueuedMessagesForSigning[JSON.stringify(params)] ?? {}
		},
				getMessagesInQueue: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.MessagesInQueue[JSON.stringify(params)] ?? {}
		},
				getGetAllQueueNames: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.GetAllQueueNames[JSON.stringify(params)] ?? {}
		},
				
		getTypeStructure: (state) => (type) => {
			return state._Structure[type].fields
		},
		getRegistry: (state) => {
			return state._Registry
		}
	},
	actions: {
		init({ dispatch, rootGetters }) {
			console.log('Vuex module: volumefi.paloma.consensus initialized!')
			if (rootGetters['common/env/client']) {
				rootGetters['common/env/client'].on('newblock', () => {
					dispatch('StoreUpdate')
				})
			}
		},
		resetState({ commit }) {
			commit('RESET_STATE')
		},
		unsubscribe({ commit }, subscription) {
			commit('UNSUBSCRIBE', subscription)
		},
		async StoreUpdate({ state, dispatch }) {
			state._Subscriptions.forEach(async (subscription) => {
				try {
					const sub=JSON.parse(subscription)
					await dispatch(sub.action, sub.payload)
				}catch(e) {
					throw new Error('Subscriptions: ' + e.message)
				}
			})
		},
		
		
		
		 		
		
		
		async QueryParams({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryParams()).data
				
					
				commit('QUERY', { query: 'Params', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryParams', payload: { options: { all }, params: {...key},query }})
				return getters['getParams']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryParams API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		
		
		 		
		
		
		async QueryQueuedMessagesForSigning({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryQueuedMessagesForSigning(query)).data
				
					
				while (all && (<any> value).pagination && (<any> value).pagination.next_key!=null) {
					let next_values=(await queryClient.queryQueuedMessagesForSigning({...query, 'pagination.key':(<any> value).pagination.next_key})).data
					value = mergeResults(value, next_values);
				}
				commit('QUERY', { query: 'QueuedMessagesForSigning', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryQueuedMessagesForSigning', payload: { options: { all }, params: {...key},query }})
				return getters['getQueuedMessagesForSigning']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryQueuedMessagesForSigning API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		
		
		 		
		
		
		async QueryMessagesInQueue({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryMessagesInQueue( key.queueTypeName)).data
				
					
				commit('QUERY', { query: 'MessagesInQueue', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryMessagesInQueue', payload: { options: { all }, params: {...key},query }})
				return getters['getMessagesInQueue']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryMessagesInQueue API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		
		
		 		
		
		
		async QueryGetAllQueueNames({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryGetAllQueueNames()).data
				
					
				commit('QUERY', { query: 'GetAllQueueNames', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryGetAllQueueNames', payload: { options: { all }, params: {...key},query }})
				return getters['getGetAllQueueNames']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryGetAllQueueNames API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		async sendMsgDeleteJob({ rootGetters }, { value, fee = [], memo = '' }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgDeleteJob(value)
				const result = await txClient.signAndBroadcast([msg], {fee: { amount: fee, 
	gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgDeleteJob:Init Could not initialize signing client. Wallet is required.')
				}else{
					throw new Error('TxClient:MsgDeleteJob:Send Could not broadcast Tx: '+ e.message)
				}
			}
		},
		async sendMsgAddMessagesSignatures({ rootGetters }, { value, fee = [], memo = '' }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgAddMessagesSignatures(value)
				const result = await txClient.signAndBroadcast([msg], {fee: { amount: fee, 
	gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgAddMessagesSignatures:Init Could not initialize signing client. Wallet is required.')
				}else{
					throw new Error('TxClient:MsgAddMessagesSignatures:Send Could not broadcast Tx: '+ e.message)
				}
			}
		},
		
		async MsgDeleteJob({ rootGetters }, { value }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgDeleteJob(value)
				return msg
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgDeleteJob:Init Could not initialize signing client. Wallet is required.')
				} else{
					throw new Error('TxClient:MsgDeleteJob:Create Could not create message: ' + e.message)
				}
			}
		},
		async MsgAddMessagesSignatures({ rootGetters }, { value }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgAddMessagesSignatures(value)
				return msg
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgAddMessagesSignatures:Init Could not initialize signing client. Wallet is required.')
				} else{
					throw new Error('TxClient:MsgAddMessagesSignatures:Create Could not create message: ' + e.message)
				}
			}
		},
		
	}
}
