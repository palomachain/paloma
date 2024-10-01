package filters

import (
	"testing"

	evmtypes "github.com/palomachain/paloma/v2/x/evm/types"
	"github.com/stretchr/testify/assert"
)

func Test_IsAssignedTo(t *testing.T) {
	msg := &evmtypes.Message{
		Assignee: "assignee",
	}
	assignee := "assignee"
	assert.True(t, IsAssignedTo(msg, assignee))
	assert.False(t, IsAssignedTo(msg, "assignee1"))
}
