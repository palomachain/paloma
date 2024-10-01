package types

import "github.com/palomachain/paloma/v2/util/blocks"

func (m *BridgeTransferLimit) BlockLimit() int64 {
	switch m.LimitPeriod {
	case LimitPeriod_DAILY:
		return blocks.DailyHeight
	case LimitPeriod_WEEKLY:
		return blocks.WeeklyHeight
	case LimitPeriod_MONTHLY:
		return blocks.MonthlyHeight
	case LimitPeriod_YEARLY:
		return blocks.YearlyHeight
	default:
		return 0
	}
}
