package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/palomachain/paloma/x/bech32ibc/types"
)

// GetFeeToken returns the fee token record for a specific denom
func (k Keeper) GetNativeHrp(ctx sdk.Context) (hrp string, err error) {
	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.NativeHrpKey) {
		return "", types.ErrNoNativeHrp
	}

	bz := store.Get(types.NativeHrpKey)

	return string(bz), nil
}

// SetNativeHrp sets the native prefix for the chain. Should only be used once.
func (k Keeper) SetNativeHrp(ctx sdk.Context, hrp string) error {
	store := ctx.KVStore(k.storeKey)

	err := types.ValidateHrp(hrp)
	if err != nil {
		return err
	}

	store.Set(types.NativeHrpKey, []byte(hrp))
	return nil
}

// GetHrpIbcRecord returns the hrp ibc record for a specific hrp
func (k Keeper) GetHrpIbcRecord(ctx sdk.Context, hrp string) (types.HrpIbcRecord, error) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.HrpIBCRecordStorePrefix)
	if !prefixStore.Has([]byte(hrp)) {
		return types.HrpIbcRecord{}, sdkerrors.Wrap(types.ErrRecordNotFound, fmt.Sprintf("hrp record not found for %s", hrp))
	}
	bz := prefixStore.Get([]byte(hrp))

	record := types.HrpIbcRecord{}
	err := proto.Unmarshal(bz, &record)
	if err != nil {
		return types.HrpIbcRecord{}, err
	}

	return record, nil
}

// setHrpIbcRecord sets a new hrp ibc record for a specific denom
func (k Keeper) setHrpIbcRecord(ctx sdk.Context, hrpIbcRecord types.HrpIbcRecord) error {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.HrpIBCRecordStorePrefix)

	if hrpIbcRecord.SourceChannel == "" {
		if prefixStore.Has([]byte(hrpIbcRecord.Hrp)) {
			prefixStore.Delete([]byte(hrpIbcRecord.Hrp))
		}
		return nil
	}

	bz, err := proto.Marshal(&hrpIbcRecord)
	if err != nil {
		return err
	}

	prefixStore.Set([]byte(hrpIbcRecord.Hrp), bz)
	return nil
}

func (k Keeper) GetHrpIbcRecords(ctx sdk.Context) (HrpIbcRecords []types.HrpIbcRecord) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.HrpIBCRecordStorePrefix)

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	records := []types.HrpIbcRecord{}

	for ; iterator.Valid(); iterator.Next() {

		record := types.HrpIbcRecord{}

		err := proto.Unmarshal(iterator.Value(), &record)
		if err != nil {
			panic(err)
		}

		records = append(records, record)
	}
	return records
}

func (k Keeper) SetHrpIbcRecords(ctx sdk.Context, hrpIbcRecords []types.HrpIbcRecord) {
	for _, record := range hrpIbcRecords {
		k.setHrpIbcRecord(ctx, record)
	}
}
