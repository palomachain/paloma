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
	"github.com/palomachain/paloma/x/paloma"
	"github.com/palomachain/paloma/x/paloma/types"
	vtypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"
)

type tx struct {
	msgs    []sdk.Msg
	address sdk.Address
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
