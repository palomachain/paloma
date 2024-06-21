package types

const (
	// dailyBlocks estimated with a block time of 1.5s
	dailyBlocks                      = 57_600
	BridgeTransferLimitDailyBlocks   = dailyBlocks
	BridgeTransferLimitWeeklyBlocks  = dailyBlocks * 7
	BridgeTransferLimitMonthlyBlocks = dailyBlocks * 30
	BridgeTransferLimitYearlyBlocks  = dailyBlocks * 365
)

func (m *BridgeTransferLimit) BlockLimit() int64 {
	switch m.LimitPeriod {
	case LimitPeriod_DAILY:
		return BridgeTransferLimitDailyBlocks
	case LimitPeriod_WEEKLY:
		return BridgeTransferLimitWeeklyBlocks
	case LimitPeriod_MONTHLY:
		return BridgeTransferLimitMonthlyBlocks
	case LimitPeriod_YEARLY:
		return BridgeTransferLimitYearlyBlocks
	default:
		return 0
	}
}
