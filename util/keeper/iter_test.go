package keeper_test

import (
	"testing"

	"github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/consensus/types"
	"github.com/stretchr/testify/assert"
)

func TestIteration(t *testing.T) {
	inData := []*types.SimpleMessage{
		{
			Sender: "1",
		},
		{
			Sender: "2",
		},
		{
			Sender: "3",
		},
	}
	ms, kv, _ := keeper.SampleStore("store", "mem")
	store := ms.GetKVStore(kv)

	for i := 0; i < len(inData); i++ {
		bz, err := types.ModuleCdc.Marshal(inData[i])
		assert.NoError(t, err)
		store.Set([]byte{byte(i)}, bz)
	}

	_, all, err := keeper.IterAll[*types.SimpleMessage](store, types.ModuleCdc)
	assert.NoError(t, err)
	assert.Len(t, all, 3)
	assert.Equal(t, inData, all)
}

func TestIterationWithError(t *testing.T) {
	ms, kv, _ := keeper.SampleStore("store", "mem")
	store := ms.GetKVStore(kv)
	store.Set([]byte("1"), []byte("something that's can't be unmarshalled"))

	_, _, err := keeper.IterAll[*types.SimpleMessage](store, types.ModuleCdc)
	assert.Error(t, err)
}
