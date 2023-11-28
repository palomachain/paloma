package keeper_test

import (
	"errors"
	"testing"

	"github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/x/consensus/types"
	"github.com/stretchr/testify/assert"
)

func Test_SaveAndLoad(t *testing.T) {
	inData := &types.SimpleMessage{
		Sender: "1",
		Hello:  "hello",
		World:  "bob",
	}
	ms, kv, _ := keeper.SampleStore("store", "mem")
	store := ms.GetKVStore(kv)

	err := keeper.Save(store, types.ModuleCdc, []byte("key"), inData)
	assert.NoError(t, err)

	ret, err := keeper.Load[*types.SimpleMessage](store, types.ModuleCdc, []byte("key"))
	assert.NoError(t, err)
	assert.Equal(t, inData, ret)
}

func Test_SaveAndLoadWithInvalidKey(t *testing.T) {
	ms, kv, _ := keeper.SampleStore("store", "mem")
	store := ms.GetKVStore(kv)
	_, err := keeper.Load[*types.SimpleMessage](store, types.ModuleCdc, []byte("i don't exist"))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, keeper.ErrNotFound))
}
