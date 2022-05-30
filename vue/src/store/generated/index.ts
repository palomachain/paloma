// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.

import PalomachainPalomaPalomachainPalomaEvm from './palomachain/paloma/palomachain.paloma.evm'
import PalomachainPalomaVolumefiPalomaConsensus from './palomachain/paloma/volumefi.paloma.consensus'
import PalomachainPalomaVolumefiPalomaScheduler from './palomachain/paloma/volumefi.paloma.scheduler'
import PalomachainPalomaVolumefiPalomaValset from './palomachain/paloma/volumefi.paloma.valset'


export default { 
  PalomachainPalomaPalomachainPalomaEvm: load(PalomachainPalomaPalomachainPalomaEvm, 'palomachain.paloma.evm'),
  PalomachainPalomaVolumefiPalomaConsensus: load(PalomachainPalomaVolumefiPalomaConsensus, 'volumefi.paloma.consensus'),
  PalomachainPalomaVolumefiPalomaScheduler: load(PalomachainPalomaVolumefiPalomaScheduler, 'volumefi.paloma.scheduler'),
  PalomachainPalomaVolumefiPalomaValset: load(PalomachainPalomaVolumefiPalomaValset, 'volumefi.paloma.valset'),
  
}


function load(mod, fullns) {
    return function init(store) {        
        if (store.hasModule([fullns])) {
            throw new Error('Duplicate module name detected: '+ fullns)
        }else{
            store.registerModule([fullns], mod)
            store.subscribe((mutation) => {
                if (mutation.type == 'common/env/INITIALIZE_WS_COMPLETE') {
                    store.dispatch(fullns+ '/init', null, {
                        root: true
                    })
                }
            })
        }
    }
}
