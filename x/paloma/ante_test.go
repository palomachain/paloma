package paloma_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/log"
	"cosmossdk.io/x/feegrant"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/x/paloma"
	"github.com/palomachain/paloma/v2/x/paloma/types"
	vtypes "github.com/palomachain/paloma/v2/x/valset/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ sdk.FeeTx = (*tx)(nil)

type tx struct {
	msgs    []sdk.Msg
	address sdk.Address
}

func (t *tx) FeeGranter() []byte {
	return nil
}

func (t *tx) FeePayer() []byte {
	return t.address.Bytes()
}

func (t *tx) GetFee() sdk.Coins {
	return nil
}

func (t *tx) GetGas() uint64 {
	return 0
}

func (t *tx) GetMsgs() []sdk.Msg {
	return t.msgs
}

func (t *tx) GetMsgsV2() ([]protov2.Message, error) {
	return []protov2.Message{&bankv1beta1.MsgSend{FromAddress: t.address.String()}}, nil // this is a hack for tests
}

func (t *tx) ValidateBasic() error {
	return nil
}

type feegrantKeeperCall struct {
	expected *feegrant.QueryAllowancesByGranterRequest
	response *feegrant.QueryAllowancesByGranterResponse
	err      error
}
type mockFeegrantKeeper struct {
	t     *testing.T
	calls []feegrantKeeperCall
	idx   int
}

func (m *mockFeegrantKeeper) addCall(req *feegrant.QueryAllowancesByGranterRequest, response *feegrant.QueryAllowancesByGranterResponse, err error) {
	if m.calls == nil {
		m.calls = make([]feegrantKeeperCall, 0)
	}

	m.calls = append(m.calls, feegrantKeeperCall{
		expected: req,
		response: response,
		err:      err,
	})
}

func (m *mockFeegrantKeeper) AllowancesByGranter(ctx context.Context, req *feegrant.QueryAllowancesByGranterRequest) (*feegrant.QueryAllowancesByGranterResponse, error) {
	if m.idx >= len(m.calls) {
		m.t.Fatalf("FAIL - FeegrantKeeper.AllowancesByGranter - only expected %d calls", m.idx)
	}

	call := m.calls[m.idx]
	if !reflect.DeepEqual(req, call.expected) {
		m.t.Fatalf("FAIL - FeegrantKeeper.AllowancesByGranter - expected: %v, got: %v", call.expected, req)
	}

	m.idx++
	return call.response, call.err
}

func (m *mockFeegrantKeeper) GrantAllowance(
	_ context.Context,
	_, _ sdk.AccAddress,
	_ feegrant.FeeAllowanceI,
) error {
	return nil
}

type mockMsg struct{}

func (m *mockMsg) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }
func (m *mockMsg) ValidateBasic() error         { return nil }
func (m *mockMsg) ProtoMessage()                {}
func (m *mockMsg) Reset()                       {}
func (m *mockMsg) String() string               { return "foo" }

func Test_VerifyAuthorisedSignatureDecorator(t *testing.T) {
	for i, tt := range []struct {
		err   error
		setup func(*mockFeegrantKeeper) sdk.Tx
		name  string
	}{
		{
			name: "with no msgs in tx",
			setup: func(*mockFeegrantKeeper) sdk.Tx {
				return &tx{}
			},
		},
		{
			name: "with msg without metadata",
			setup: func(*mockFeegrantKeeper) sdk.Tx {
				return &tx{
					msgs: []sdk.Msg{
						&mockMsg{},
					},
				}
			},
		},
		{
			name: "with msg signed by creator",
			setup: func(*mockFeegrantKeeper) sdk.Tx {
				valAddr := sdk.AccAddress("foo")
				return &tx{
					msgs: []sdk.Msg{
						&types.MsgAddStatusUpdate{
							Status: "bar",
							Level:  0,
							Metadata: vtypes.MsgMetadata{
								Creator: valAddr.String(),
								Signers: []string{valAddr.String()},
							},
						},
					},
				}
			},
		},
		{
			name: "with msg signed by valid signature",
			setup: func(k *mockFeegrantKeeper) sdk.Tx {
				valAddr := sdk.AccAddress("foo")
				grantee := sdk.AccAddress("bar")
				k.addCall(&feegrant.QueryAllowancesByGranterRequest{Granter: valAddr.String()},
					&feegrant.QueryAllowancesByGranterResponse{
						Allowances: []*feegrant.Grant{
							{
								Granter: valAddr.String(),
								Grantee: grantee.String(),
							},
						},
					},
					nil)
				return &tx{
					msgs: []sdk.Msg{
						&types.MsgAddStatusUpdate{
							Status: "bar",
							Level:  0,
							Metadata: vtypes.MsgMetadata{
								Creator: valAddr.String(),
								Signers: []string{grantee.String()},
							},
						},
					},
				}
			},
		},
		{
			name: "with msg signed by multiple valid signatures",
			setup: func(k *mockFeegrantKeeper) sdk.Tx {
				valAddr := sdk.AccAddress("foo")
				grantee := sdk.AccAddress("bar")
				grantee2 := sdk.AccAddress("bar")
				k.addCall(&feegrant.QueryAllowancesByGranterRequest{Granter: valAddr.String()},
					&feegrant.QueryAllowancesByGranterResponse{
						Allowances: []*feegrant.Grant{
							{
								Granter: valAddr.String(),
								Grantee: grantee.String(),
							},
							{
								Granter: valAddr.String(),
								Grantee: grantee2.String(),
							},
						},
					},
					nil)
				return &tx{
					msgs: []sdk.Msg{
						&types.MsgAddStatusUpdate{
							Status: "bar",
							Level:  0,
							Metadata: vtypes.MsgMetadata{
								Creator: valAddr.String(),
								Signers: []string{grantee.String(), grantee2.String()},
							},
						},
					},
				}
			},
		},
		{
			name: "with msg signed by multiple signatures, including at least one valid signature",
			setup: func(k *mockFeegrantKeeper) sdk.Tx {
				valAddr := sdk.AccAddress("foo")
				grantee := sdk.AccAddress("bar")
				grantee2 := sdk.AccAddress("bar")
				k.addCall(&feegrant.QueryAllowancesByGranterRequest{Granter: valAddr.String()},
					&feegrant.QueryAllowancesByGranterResponse{
						Allowances: []*feegrant.Grant{
							{
								Granter: valAddr.String(),
								Grantee: grantee.String(),
							},
						},
					},
					nil)
				return &tx{
					msgs: []sdk.Msg{
						&types.MsgAddStatusUpdate{
							Status: "bar",
							Level:  0,
							Metadata: vtypes.MsgMetadata{
								Creator: valAddr.String(),
								Signers: []string{grantee.String(), grantee2.String()},
							},
						},
					},
				}
			},
		},
		{
			name: "with msg signed by multiple signatures without at least one valid signature",
			err:  fmt.Errorf("no signature from granted address found for message"),
			setup: func(k *mockFeegrantKeeper) sdk.Tx {
				valAddr := sdk.AccAddress("foo")
				grantee := sdk.AccAddress("bar")
				grantee2 := sdk.AccAddress("bar")
				k.addCall(&feegrant.QueryAllowancesByGranterRequest{Granter: valAddr.String()},
					&feegrant.QueryAllowancesByGranterResponse{
						Allowances: []*feegrant.Grant{},
					},
					nil)
				return &tx{
					msgs: []sdk.Msg{
						&types.MsgAddStatusUpdate{
							Status: "bar",
							Level:  0,
							Metadata: vtypes.MsgMetadata{
								Creator: valAddr.String(),
								Signers: []string{grantee.String(), grantee2.String()},
							},
						},
					},
				}
			},
		},
	} {
		t.Run(fmt.Sprintf("%d. %s", i, tt.name), func(t *testing.T) {
			r := require.New(t)
			k := &mockFeegrantKeeper{t: t}
			tx := tt.setup(k)
			testee := paloma.NewVerifyAuthorisedSignatureDecorator(k)
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
			_, err := testee.AnteHandle(ctx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) { return ctx, nil })

			if tt.err != nil {
				r.ErrorContains(err, tt.err.Error())
			} else {
				r.NoError(err)
			}
			r.Len(k.calls, k.idx, "unexpected amount of calls")
		})
	}
}

func Test_TxFeeSkipper(t *testing.T) {
	t.Run("with no msgs in tx", func(t *testing.T) {
		ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
		tx := &tx{
			msgs: []sdk.Msg{
				&types.MsgAddStatusUpdate{
					Status: "bar",
					Level:  0,
					Metadata: vtypes.MsgMetadata{
						Creator: "foo",
						Signers: []string{"foo"},
					},
				},
			},
		}
		fee, p, err := paloma.TxFeeSkipper(ctx, tx)

		require.Equal(t, sdk.Coins{}, fee)
		require.True(t, fee.IsZero())
		require.Equal(t, int64(42), p)
		require.NoError(t, err)
	})

	t.Run("with nil tx", func(t *testing.T) {
		ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
		fee, p, err := paloma.TxFeeSkipper(ctx, nil)

		require.Equal(t, sdk.Coins{}, fee)
		require.True(t, fee.IsZero())
		require.Equal(t, int64(42), p)
		require.NoError(t, err)
	})
}

func Test_GasExemptAddressDecorator(t *testing.T) {
	t.Run("with gas exempt address", func(t *testing.T) {
		ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
		sender := sdk.AccAddress("exempt")
		tx := &tx{
			address: sender,
		}
		params := types.Params{
			GasExemptAddresses: []string{sender.String()},
		}
		k := &mockPalomaKeeper{params: params}
		decorator := paloma.NewGasExemptAddressDecorator(k)

		_, err := decorator.AnteHandle(ctx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
			require.Contains(t, ctx.GasMeter().String(), "freeGasMeter")
			return ctx, nil
		})

		require.NoError(t, err)
	})

	t.Run("without gas exempt address", func(t *testing.T) {
		ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
		tx := &tx{
			address: sdk.AccAddress("non-exempt"),
		}
		params := types.Params{
			GasExemptAddresses: []string{sdk.AccAddress("exempt").String()},
		}
		k := &mockPalomaKeeper{params: params}
		decorator := paloma.NewGasExemptAddressDecorator(k)

		_, err := decorator.AnteHandle(ctx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
			require.Contains(t, ctx.GasMeter().String(), "InfiniteGasMeter")
			return ctx, nil
		})

		require.NoError(t, err)
	})

	t.Run("with invalid transaction type", func(t *testing.T) {
		ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
		tx := &mockInvalidTx{}
		params := types.Params{
			GasExemptAddresses: []string{"exempt"},
		}
		k := &mockPalomaKeeper{params: params}
		decorator := paloma.NewGasExemptAddressDecorator(k)

		_, err := decorator.AnteHandle(ctx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
			return ctx, nil
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid transaction type")
	})
}

type mockPalomaKeeper struct {
	params types.Params
}

func (m *mockPalomaKeeper) GetParams(ctx context.Context) types.Params {
	return m.params
}

type mockInvalidTx struct{}

func (m *mockInvalidTx) GetMsgsV2() ([]protoreflect.ProtoMessage, error) {
	return nil, nil
}

func (m *mockInvalidTx) GetMsgs() []sdk.Msg {
	return nil
}
