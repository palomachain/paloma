import { txClient, queryClient, MissingWalletError , registry} from './module'

import { Params } from "./module/types/valset/params"
import { Validator } from "./module/types/valset/snapshot"
import { Snapshot } from "./module/types/valset/snapshot"
import { ExternalChainInfo } from "./module/types/valset/snapshot"
import { MsgAddExternalChainInfoForValidator_ChainInfo } from "./module/types/valset/tx"


export { Params, Validator, Snapshot, ExternalChainInfo, MsgAddExternalChainInfoForValidator_ChainInfo };

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
				ValidatorInfo: {},
				
				_Structure: {
						Params: getStructure(Params.fromPartial({})),
						Validator: getStructure(Validator.fromPartial({})),
						Snapshot: getStructure(Snapshot.fromPartial({})),
						ExternalChainInfo: getStructure(ExternalChainInfo.fromPartial({})),
						MsgAddExternalChainInfoForValidator_ChainInfo: getStructure(MsgAddExternalChainInfoForValidator_ChainInfo.fromPartial({})),
						
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
				getValidatorInfo: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.ValidatorInfo[JSON.stringify(params)] ?? {}
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
			console.log('Vuex module: volumefi.paloma.valset initialized!')
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
		
		
		
		
		 		
		
		
		async QueryValidatorInfo({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryValidatorInfo( key.valAddr)).data
				
					
				commit('QUERY', { query: 'ValidatorInfo', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryValidatorInfo', payload: { options: { all }, params: {...key},query }})
				return getters['getValidatorInfo']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryValidatorInfo API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		async sendMsgAddExternalChainInfoForValidator({ rootGetters }, { value, fee = [], memo = '' }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgAddExternalChainInfoForValidator(value)
				const result = await txClient.signAndBroadcast([msg], {fee: { amount: fee, 
	gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgAddExternalChainInfoForValidator:Init Could not initialize signing client. Wallet is required.')
				}else{
					throw new Error('TxClient:MsgAddExternalChainInfoForValidator:Send Could not broadcast Tx: '+ e.message)
				}
			}
		},
		async sendMsgRegisterConductor({ rootGetters }, { value, fee = [], memo = '' }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgRegisterConductor(value)
				const result = await txClient.signAndBroadcast([msg], {fee: { amount: fee, 
	gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgRegisterConductor:Init Could not initialize signing client. Wallet is required.')
				}else{
					throw new Error('TxClient:MsgRegisterConductor:Send Could not broadcast Tx: '+ e.message)
				}
			}
		},
		
		async MsgAddExternalChainInfoForValidator({ rootGetters }, { value }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgAddExternalChainInfoForValidator(value)
				return msg
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgAddExternalChainInfoForValidator:Init Could not initialize signing client. Wallet is required.')
				} else{
					throw new Error('TxClient:MsgAddExternalChainInfoForValidator:Create Could not create message: ' + e.message)
				}
			}
		},
		async MsgRegisterConductor({ rootGetters }, { value }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgRegisterConductor(value)
				return msg
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgRegisterConductor:Init Could not initialize signing client. Wallet is required.')
				} else{
					throw new Error('TxClient:MsgRegisterConductor:Create Could not create message: ' + e.message)
				}
			}
		},
		
	}
}
