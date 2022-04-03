// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.

import VolumefiCronchainVolumefiCronchainScheduler from './volumefi/cronchain/volumefi.cronchain.scheduler'


export default { 
  VolumefiCronchainVolumefiCronchainScheduler: load(VolumefiCronchainVolumefiCronchainScheduler, 'volumefi.cronchain.scheduler'),
  
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
