package client

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/assert"
	testdata_type "github.com/volumefi/cronchain/runner/testdata/types"
)

type testAccountInfo struct {
	uid      string
	mnemonic string
	password string
}

func (acc testAccountInfo) addToKeybase(kb keyring.Keyring) keyring.Info {
	hdpath := hd.NewFundraiserParams(0, sdk.CoinType, 0).String()
	info, err := kb.NewAccount(acc.uid, acc.mnemonic, acc.password, hdpath, hd.Secp256k1)
	if err != nil {
		panic(err)
	}
	return info
}

var (
	bob = testAccountInfo{
		uid:      "bob",
		mnemonic: "brush tattoo leave reunion list expand crime wage festival brass foster giggle tumble rice hybrid else general dwarf supreme wave select riot fault speak",
	}
	alice = testAccountInfo{
		uid:      "alice",
		mnemonic: "torch cabin vanish scrub wrong one piano side hurt frozen country empower steak erase oblige elder direct feel grant village follow abuse basket display",
	}
)

type fakeAccountFetcher func(ctx context.Context, address sdk.AccAddress) (authtypes.AccountI, error)

func (f fakeAccountFetcher) QueryAccount(ctx context.Context, address sdk.AccAddress) (authtypes.AccountI, error) {
	return f(ctx, address)
}

func mockAccFetcher(acc *authtypes.BaseAccount, err error) fakeAccountFetcher {
	return fakeAccountFetcher(func(ctx context.Context, address sdk.AccAddress) (authtypes.AccountI, error) {
		return acc, err
	})
}

func TestMessageSigningAndSignatureVerification(t *testing.T) {
	var (
		c   *LensClient = newClient()
		msg testdata_type.SimpleMessage
	)
	c.Keybase = keyring.NewInMemory()
	bobInfo := bob.addToKeybase(c.Keybase)
	aliceInfo := alice.addToKeybase(c.Keybase)

	c.Config.Key = bob.uid
	c.Config.AccountPrefix = "cosmos"

	msg.Sender = bobInfo.GetAddress().String()
	msg.Hello = "hey"
	msg.World = "mars"

	tx, err := c.OfflineSignMsgs(123, 456, &msg)
	assert.NoError(t, err)
	signatures, err := tx.GetSignaturesV2()
	assert.NoError(t, err)
	assert.Len(t, signatures, 1)

	for _, tt := range []struct {
		name        string
		addr        sdk.AccAddress
		pub         cryptotypes.PubKey
		accNum      uint64
		seq         uint64
		expectedErr error
	}{
		{
			"signed and verified by same acc",
			bobInfo.GetAddress(),
			bobInfo.GetPubKey(),
			123,
			456,
			nil,
		},
		{
			"the seq does not matter for sig verification",
			bobInfo.GetAddress(),
			bobInfo.GetPubKey(),
			123,
			0xBABA,
			nil,
		},
		{
			"if the account number is different then it fails",
			bobInfo.GetAddress(),
			bobInfo.GetPubKey(),
			1,
			0xBABA,
			ErrSignatureVerificationFailed,
		},
		{
			"signed and verified by other acc returns an error",
			aliceInfo.GetAddress(),
			aliceInfo.GetPubKey(),
			111,
			999,
			ErrSignatureVerificationFailed,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(
				t,
				verifySignature(
					context.Background(),
					c.Config.ChainID,
					mockAccFetcher(
						// address and public key don't matter.
						// only thing that does matter is the acc num and seq
						authtypes.NewBaseAccount(tt.addr, tt.pub, tt.accNum, tt.seq),
						nil,
					),
					c.Codec.TxConfig.SignModeHandler(),
					tx,
				),
				tt.expectedErr,
			)
		})

	}

}
