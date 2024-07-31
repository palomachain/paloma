package types

import (
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeepAliveData struct {
	ContactedAt           time.Time
	ValAddr               sdk.ValAddress
	AliveUntilBlockHeight int64
	PigeonVersion         string
}
