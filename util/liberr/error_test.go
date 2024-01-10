package liberr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/palomachain/paloma/util/liberr"
	"github.com/stretchr/testify/require"
)

const dummyErr = liberr.Error("not found: %s")

func TestFormat(t *testing.T) {
	e := dummyErr.Format("foo")

	fmt.Println(errors.Is(e, dummyErr))
	require.True(t, errors.Is(e, dummyErr))

	e = e.JoinErrorf("with extra info: %s", "bar")
	require.True(t, errors.Is(e, dummyErr))
}
