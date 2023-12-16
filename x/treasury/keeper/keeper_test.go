package keeper

import (
	"errors"
	"testing"

	"cosmossdk.io/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	keeperutilmocks "github.com/palomachain/paloma/util/keeper/mocks"
	"github.com/palomachain/paloma/x/treasury/types"
	"github.com/palomachain/paloma/x/treasury/types/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetCommunityFundFee(t *testing.T) {
	testcases := []struct {
		name        string
		setup       func() Keeper
		input       string
		expectedErr error
	}{
		{
			name:  "success case - tells store to set fees and returns no error",
			input: "0.01",
			setup: func() Keeper {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)
				keeperUtil.On("Load",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&types.Fees{}, nil)

				keeperUtil.On("Save",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					&types.Fees{
						CommunityFundFee: "0.01",
					},
				).Return(nil)

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k
			},
		},
		{
			name:  "success case with existing fees - tells store to set fees and returns no error",
			input: "0.01",
			setup: func() Keeper {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)
				keeperUtil.On("Load",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&types.Fees{
					CommunityFundFee: "0.02",
					SecurityFee:      "0.03",
				}, nil)

				keeperUtil.On("Save",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					&types.Fees{
						CommunityFundFee: "0.01",
						SecurityFee:      "0.03",
					},
				).Return(nil)

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k
			},
		},
		{
			name:  "error returned loading existing fees, returns error",
			input: "0.01",
			setup: func() Keeper {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)
				keeperUtil.On("Load",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&types.Fees{}, errors.New("load error"))

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k
			},
			expectedErr: errors.New("load error"),
		},
		{
			name:  "error returned saving fees, returns error",
			input: "0.01",
			setup: func() Keeper {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)
				keeperUtil.On("Load",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&types.Fees{}, nil)

				keeperUtil.On("Save",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					&types.Fees{
						CommunityFundFee: "0.01",
					},
				).Return(errors.New("save error"))

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k
			},
			expectedErr: errors.New("save error"),
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
			k := tt.setup()

			actualErr := k.SetCommunityFundFee(ctx, tt.input)
			asserter.Equal(tt.expectedErr, actualErr)
		})
	}
}

func TestSetSecurityFee(t *testing.T) {
	testcases := []struct {
		name        string
		setup       func() Keeper
		input       string
		expectedErr error
	}{
		{
			name:  "success case - tells store to set fees and returns no error",
			input: "0.01",
			setup: func() Keeper {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)
				keeperUtil.On("Load",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&types.Fees{}, nil)

				keeperUtil.On("Save",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					&types.Fees{
						SecurityFee: "0.01",
					},
				).Return(nil)

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k
			},
		},
		{
			name:  "success case with existing fees - tells store to set fees and returns no error",
			input: "0.01",
			setup: func() Keeper {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)
				keeperUtil.On("Load",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&types.Fees{
					CommunityFundFee: "0.02",
					SecurityFee:      "0.03",
				}, nil)

				keeperUtil.On("Save",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					&types.Fees{
						CommunityFundFee: "0.02",
						SecurityFee:      "0.01",
					},
				).Return(nil)

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k
			},
		},
		{
			name:  "error returned loading existing fees, returns error",
			input: "0.01",
			setup: func() Keeper {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)
				keeperUtil.On("Load",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&types.Fees{}, errors.New("load error"))

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k
			},
			expectedErr: errors.New("load error"),
		},
		{
			name:  "error returned saving fees, returns error",
			input: "0.01",
			setup: func() Keeper {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)
				keeperUtil.On("Load",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&types.Fees{}, nil)

				keeperUtil.On("Save",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					&types.Fees{
						SecurityFee: "0.01",
					},
				).Return(errors.New("save error"))

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k
			},
			expectedErr: errors.New("save error"),
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
			k := tt.setup()

			actualErr := k.SetSecurityFee(ctx, tt.input)
			asserter.Equal(tt.expectedErr, actualErr)
		})
	}
}

func TestGetFees(t *testing.T) {
	testcases := []struct {
		name        string
		setup       func() Keeper
		expected    *types.Fees
		expectedErr error
	}{
		{
			name: "success case - returns what's loaded from the store",
			setup: func() Keeper {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)
				keeperUtil.On("Load",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&types.Fees{
					CommunityFundFee: "0.01",
					SecurityFee:      "0.02",
				}, nil)

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k
			},
			expected: &types.Fees{
				CommunityFundFee: "0.01",
				SecurityFee:      "0.02",
			},
		},
		{
			name: "success case - not found in store.  returns empty fees",
			setup: func() Keeper {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)
				keeperUtil.On("Load",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&types.Fees{}, keeperutil.ErrNotFound.Format(&types.Fees{}, ""))

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k
			},
			expected: &types.Fees{},
		},
		{
			name: "error case - returns error from loading",
			setup: func() Keeper {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)
				keeperUtil.On("Load",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&types.Fees{}, errors.New("load error"))

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k
			},
			expected:    &types.Fees{},
			expectedErr: errors.New("load error"),
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
			k := tt.setup()

			actual, actualErr := k.GetFees(ctx)
			asserter.Equal(tt.expected, actual)
			asserter.Equal(tt.expectedErr, actualErr)
		})
	}
}

func TestSetFees(t *testing.T) {
	testcases := []struct {
		name        string
		setup       func() Keeper
		input       *types.Fees
		expectedErr error
	}{
		{
			name: "success case - tells store to set fees and returns no error",
			input: &types.Fees{
				CommunityFundFee: "0.01",
				SecurityFee:      "0.02",
			},
			setup: func() Keeper {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)

				keeperUtil.On("Save",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					&types.Fees{
						CommunityFundFee: "0.01",
						SecurityFee:      "0.02",
					},
				).Return(nil)

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k
			},
		},
		{
			name: "error case - returns store error",
			input: &types.Fees{
				CommunityFundFee: "0.01",
				SecurityFee:      "0.02",
			},
			setup: func() Keeper {
				keeperUtil := keeperutilmocks.NewKeeperUtilI[*types.Fees](t)

				keeperUtil.On("Save",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					&types.Fees{
						CommunityFundFee: "0.01",
						SecurityFee:      "0.02",
					},
				).Return(errors.New("save error"))

				store := mocks.NewTreasuryStore(t)
				store.On("TreasuryStore", mock.Anything).Return(nil)

				k := Keeper{
					KeeperUtil: keeperUtil,
					Store:      store,
				}

				return k
			},
			expectedErr: errors.New("save error"),
		},
	}

	asserter := assert.New(t)
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
			k := tt.setup()

			actualErr := k.setFees(ctx, tt.input)
			asserter.Equal(tt.expectedErr, actualErr)
		})
	}
}
