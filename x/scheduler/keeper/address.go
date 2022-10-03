package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/palomachain/paloma/x/scheduler/types"
)

const sizeofInt = 8

func BuildAddress(blockHeight uint64, jobID string) sdk.AccAddress {
	addrbz := make([]byte, sizeofInt+sizeofInt+types.JobIDMaxLen)
	binary.BigEndian.PutUint64(addrbz[:8], uint64(len(jobID)))
	binary.BigEndian.PutUint64(addrbz[8:16], blockHeight)
	copy(addrbz[16:], []byte(jobID))
	return address.Module(types.ModuleName, addrbz)[:types.JobAddressLength]
}
