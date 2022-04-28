package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/palomachain/paloma/x/consensus/testdata/types"
)

func TestSaveAndLoad(t *testing.T) {
	inData := &types.SimpleMessage{
		Sender: "1",
		Hello:  "hello",
		World:  "bob",
	}
	ms, kv, _ := SampleStore("store", "mem")
	store := ms.GetKVStore(kv)

	err := Save(store, types.ModuleCdc, []byte("key"), inData)
	assert.NoError(t, err)

	ret, err := Load[*types.SimpleMessage](store, types.ModuleCdc, []byte("key"))
	assert.NoError(t, err)
	assert.Equal(t, inData, ret)
}

func TestSaveAndLoadWithInvalidKey(t *testing.T) {
	ms, kv, _ := SampleStore("store", "mem")
	store := ms.GetKVStore(kv)
	_, err := Load[*types.SimpleMessage](store, types.ModuleCdc, []byte("i don't exist"))
	assert.Error(t, err)
}
