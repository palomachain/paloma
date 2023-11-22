package keeper_test

import (
	"testing"

	"github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/consensus/types"
	"github.com/stretchr/testify/assert"
)

func TestSaveAndLoad(t *testing.T) {
	inData := &types.SimpleMessage{
		Sender: "1",
		Hello:  "hello",
		World:  "bob",
	}
	ms, kv, _ := keeper.SampleStore("store", "mem")
	store := ms.GetKVStore(kv)

	err := keeper.Save(store, nil, []byte("key"), inData)
	assert.NoError(t, err)

	ret, err := keeper.Load[*types.SimpleMessage](store, nil, []byte("key"))
	assert.NoError(t, err)
	assert.Equal(t, inData, ret)
}

func TestSaveAndLoadWithInvalidKey(t *testing.T) {
	ms, kv, _ := keeper.SampleStore("store", "mem")
	store := ms.GetKVStore(kv)
	_, err := keeper.Load[*types.SimpleMessage](store, nil, []byte("i don't exist"))
	assert.Error(t, err)
}
