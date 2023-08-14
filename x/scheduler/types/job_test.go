package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJobIsTargetedAtChainWithMEVRelayingSupport(t *testing.T) {
	t.Run("for chains with MEV relaying support", func(t *testing.T) {
		for _, v := range []string{"eth-main", "bnb-main", "matic-main"} {
			r := (&Job{Routing: Routing{ChainReferenceID: v}}).isTargetedAtChainWithMEVRelayingSupport()
			require.True(t, r, "chain %s should be supported", v)
		}
	})
	t.Run("for chains without MEV relaying support", func(t *testing.T) {
		for _, v := range []string{"op-main", "kava-main", "ropsen"} {
			r := (&Job{Routing: Routing{ChainReferenceID: v}}).isTargetedAtChainWithMEVRelayingSupport()
			require.False(t, r, "should return false")
		}
	})
}
