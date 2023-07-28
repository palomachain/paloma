package types

import (
	"fmt"
	math "math"
	"math/big"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

//////////////////////////////////////
/////// BRIDGE VALIDATOR(S) //////////
//////////////////////////////////////

// ToInternal transforms a BridgeValidator into its fully validated internal type
func (b BridgeValidator) ToInternal() (*InternalBridgeValidator, error) {
	return NewInternalBridgeValidator(b)
}

// BridgeValidators is the sorted set of validator data for Ethereum bridge MultiSig set
type BridgeValidators []BridgeValidator

func (b BridgeValidators) ToInternal() (*InternalBridgeValidators, error) {
	ret := make(InternalBridgeValidators, len(b))
	for i := range b {
		ibv, err := NewInternalBridgeValidator(b[i])
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "member %d", i)
		}
		ret[i] = ibv
	}
	return &ret, nil
}

// Equal checks that slice contents and order are equal
func (b BridgeValidators) Equal(o BridgeValidators) bool {
	if len(b) != len(o) {
		return false
	}

	for i, bv := range b {
		ov := o[i]
		if bv != ov {
			return false
		}
	}

	return true
}

// InternalBridgeValidator is a BridgeValidator but with validated EthereumAddress
type InternalBridgeValidator struct {
	Power           uint64
	EthereumAddress EthAddress
}

func NewInternalBridgeValidator(bridgeValidator BridgeValidator) (*InternalBridgeValidator, error) {
	ethAddr, err := NewEthAddress(bridgeValidator.EthereumAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid bridge validator eth address")
	}

	i := &InternalBridgeValidator{
		Power:           bridgeValidator.Power,
		EthereumAddress: *ethAddr,
	}
	if err := i.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid bridge validator")
	}
	return i, nil
}

func (i InternalBridgeValidator) ValidateBasic() error {
	if err := i.EthereumAddress.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "ethereum address")
	}
	return nil
}

func (i InternalBridgeValidator) ToExternal() BridgeValidator {
	return BridgeValidator{
		Power:           i.Power,
		EthereumAddress: i.EthereumAddress.GetAddress().Hex(),
	}
}

// InternalBridgeValidators is the sorted set of validator data for Ethereum bridge MultiSig set
type InternalBridgeValidators []*InternalBridgeValidator

func (i InternalBridgeValidators) ToExternal() BridgeValidators {
	bridgeValidators := make([]BridgeValidator, len(i))
	for b := range bridgeValidators {
		bridgeValidators[b] = i[b].ToExternal()
	}

	return BridgeValidators(bridgeValidators)
}

// Sort sorts the validators by power
func (b InternalBridgeValidators) Sort() {
	sort.Slice(b, func(i, j int) bool {
		if b[i].Power == b[j].Power {
			// Secondary sort on eth address in case powers are equal
			return EthAddrLessThan(b[i].EthereumAddress, b[j].EthereumAddress)
		}
		return b[i].Power > b[j].Power
	})
}

// PowerDiff returns the difference in power between two bridge validator sets
// note this is Gravity bridge power *not* Cosmos voting power. Cosmos voting
// power is based on the absolute number of tokens in the staking pool at any given
// time Gravity bridge power is normalized using the equation.
//
// validators cosmos voting power / total cosmos voting power in this block = gravity bridge power / u32_max
//
// As an example if someone has 52% of the Cosmos voting power when a validator set is created their Gravity
// bridge voting power is u32_max * .52
//
// Normalized voting power dramatically reduces how often we have to produce new validator set updates. For example
// if the total on chain voting power increases by 1% due to inflation, we shouldn't have to generate a new validator
// set, after all the validators retained their relative percentages during inflation and normalized Gravity bridge power
// shows no difference.
func (b InternalBridgeValidators) PowerDiff(c InternalBridgeValidators) float64 {
	powers := map[string]int64{}
	// loop over b and initialize the map with their powers
	for _, bv := range b {
		powers[bv.EthereumAddress.GetAddress().Hex()] = int64(bv.Power)
	}

	// subtract c powers from powers in the map, initializing
	// uninitialized keys with negative numbers
	for _, bv := range c {
		if val, ok := powers[bv.EthereumAddress.GetAddress().Hex()]; ok {
			powers[bv.EthereumAddress.GetAddress().Hex()] = val - int64(bv.Power)
		} else {
			powers[bv.EthereumAddress.GetAddress().Hex()] = -int64(bv.Power)
		}
	}

	var delta float64
	for _, v := range powers {
		// NOTE: we care about the absolute value of the changes
		delta += math.Abs(float64(v))
	}

	return math.Abs(delta / float64(math.MaxUint32))
}

// TotalPower returns the total power in the bridge validator set
func (b InternalBridgeValidators) TotalPower() (out uint64) {
	for _, v := range b {
		out += v.Power
	}
	return
}

// HasDuplicates returns true if there are duplicates in the set
func (b InternalBridgeValidators) HasDuplicates() bool {
	m := make(map[string]struct{}, len(b))
	// creates a hashmap then ensures that the hashmap and the array
	// have the same length, this acts as an O(n) duplicates check
	for i := range b {
		m[b[i].EthereumAddress.GetAddress().Hex()] = struct{}{}
	}
	return len(m) != len(b)
}

// GetPowers returns only the power values for all members
func (b InternalBridgeValidators) GetPowers() []uint64 {
	r := make([]uint64, len(b))
	for i := range b {
		r[i] = b[i].Power
	}
	return r
}

// ValidateBasic performs stateless checks
func (b InternalBridgeValidators) ValidateBasic() error {
	if len(b) == 0 {
		return ErrEmpty
	}
	for i := range b {
		if err := b[i].ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(err, "member %d", i)
		}
	}
	if b.HasDuplicates() {
		return sdkerrors.Wrap(ErrDuplicate, "addresses")
	}
	return nil
}

//////////////////////////////////////
// VALSETS              //
//////////////////////////////////////

// NewValset returns a new valset
func NewValset(nonce, height uint64, members InternalBridgeValidators, rewardAmount sdk.Int, rewardToken EthAddress) (*Valset, error) {
	if err := members.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid members")
	}
	members.Sort()
	var mem []BridgeValidator
	for _, val := range members {
		mem = append(mem, val.ToExternal())
	}
	vs := Valset{Nonce: uint64(nonce), Members: mem, Height: height, RewardAmount: rewardAmount, RewardToken: rewardToken.GetAddress().Hex()}
	return &vs,
		nil
}

// GetCheckpoint returns the checkpoint
func (v Valset) GetCheckpoint(gravityIDstring string) []byte {

	// error case here should not occur outside of testing since the above is a constant
	contractAbi, abiErr := abi.JSON(strings.NewReader(ValsetCheckpointABIJSON))
	if abiErr != nil {
		panic("Bad ABI constant!")
	}

	// the contract argument is not a arbitrary length array but a fixed length 32 byte
	// array, therefore we have to utf8 encode the string (the default in this case) and
	// then copy the variable length encoded data into a fixed length array. This function
	// will panic if gravityId is too long to fit in 32 bytes
	gravityID, err := strToFixByteArray(gravityIDstring)
	if err != nil {
		panic(err)
	}

	// this should never happen, unless an invalid parameter value has been set by the chain
	err = ValidateEthAddress(v.RewardToken)
	if err != nil {
		panic(err)
	}
	rewardToken := gethcommon.HexToAddress(v.RewardToken)

	if v.RewardAmount.BigInt() == nil {
		// this must be programmer error
		panic("Invalid reward amount passed in valset GetCheckpoint!")
	}
	rewardAmount := v.RewardAmount.BigInt()

	checkpointBytes := []uint8("checkpoint")
	var checkpoint [32]uint8
	copy(checkpoint[:], checkpointBytes)

	memberAddresses := make([]gethcommon.Address, len(v.Members))
	convertedPowers := make([]*big.Int, len(v.Members))
	for i, m := range v.Members {
		memberAddresses[i] = gethcommon.HexToAddress(m.EthereumAddress)
		convertedPowers[i] = big.NewInt(int64(m.Power))
	}
	// the word 'checkpoint' needs to be the same as the 'name' above in the checkpointAbiJson
	// but other than that it's a constant that has no impact on the output. This is because
	// it gets encoded as a function name which we must then discard.
	bytes, packErr := contractAbi.Pack("checkpoint", gravityID, checkpoint, big.NewInt(int64(v.Nonce)), memberAddresses, convertedPowers, rewardAmount, rewardToken)

	// this should never happen outside of test since any case that could crash on encoding
	// should be filtered above.
	if packErr != nil {
		panic(fmt.Sprintf("Error packing checkpoint! %s/n", packErr))
	}

	// we hash the resulting encoded bytes discarding the first 4 bytes these 4 bytes are the constant
	// method name 'checkpoint'. If you were to replace the checkpoint constant in this code you would
	// then need to adjust how many bytes you truncate off the front to get the output of abi.encode()
	hash := crypto.Keccak256Hash(bytes[4:])
	return hash.Bytes()
}

// WithoutEmptyMembers returns a new Valset without member that have 0 power or an empty Ethereum address.
func (v *Valset) WithoutEmptyMembers() *Valset {
	if v == nil {
		return nil
	}
	r := Valset{
		Nonce:        v.Nonce,
		Members:      make([]BridgeValidator, 0, len(v.Members)),
		Height:       0,
		RewardAmount: sdk.Int{},
		RewardToken:  "",
	}
	for i := range v.Members {
		if _, err := v.Members[i].ToInternal(); err == nil {
			r.Members = append(r.Members, v.Members[i])
		}
	}
	return &r
}

// Equal compares all of the valset members, additionally returning an error explaining the problem
func (v Valset) Equal(o Valset) (bool, error) {
	if v.Height != o.Height {
		return false, sdkerrors.Wrap(ErrInvalid, "valset heights mismatch")
	}

	if v.Nonce != o.Nonce {
		return false, sdkerrors.Wrap(ErrInvalid, "valset nonces mismatch")
	}

	if !v.RewardAmount.Equal(o.RewardAmount) {
		return false, sdkerrors.Wrap(ErrInvalid, "valset reward amounts mismatch")
	}

	if v.RewardToken != o.RewardToken {
		return false, sdkerrors.Wrap(ErrInvalid, "valset reward tokens mismatch")
	}

	var bvs BridgeValidators = v.Members
	var ovs BridgeValidators = o.Members
	if !bvs.Equal(ovs) {
		return false, sdkerrors.Wrap(ErrInvalid, "valset members mismatch")
	}

	return true, nil
}

func (v Valset) ValidateBasic() error {
	if len(v.Members) == 0 {
		return sdkerrors.Wrap(ErrInvalidValset, "valset must have members")
	}
	for _, mem := range v.Members {
		_, err := mem.ToInternal() // ToInternal validates the InternalBridgeValidator for us
		if err != nil {
			return err
		}
	}
	if v.RewardAmount.IsNegative() {
		return sdkerrors.Wrap(ErrInvalidValset, "valset reward must not be negative")
	}
	cleanToken := strings.TrimSpace(v.RewardToken)
	if v.RewardAmount.IsPositive() && cleanToken == "" {
		return sdkerrors.Wrap(ErrInvalidValset, "no valset reward denom for nonzero reward")
	}
	if cleanToken != v.RewardToken {
		return sdkerrors.Wrap(ErrInvalidValset, "reward token should be properly formatted")
	}
	return nil
}

// Valsets is a collection of valset
type Valsets []Valset

func (v Valsets) Len() int {
	return len(v)
}

func (v Valsets) Less(i, j int) bool {
	return v[i].Nonce > v[j].Nonce
}

func (v Valsets) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v Valsets) ValidateBasic() error {
	for _, valset := range v {
		if err := valset.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}

// GetFees returns the total fees contained within a given batch
func (b OutgoingTxBatch) GetFees() sdk.Int {
	sum := sdk.ZeroInt()
	for _, t := range b.Transactions {
		sum = sum.Add(t.Erc20Fee.Amount)
	}
	return sum
}

// This interface is implemented by all the types that are used
// to create transactions on Ethereum that are signed by validators.
// The naming here could be improved.
type EthereumSigned interface {
	GetCheckpoint(gravityIDstring string) []byte
}

// nolint: exhaustruct
var (
	_ EthereumSigned = &Valset{}
	_ EthereumSigned = &OutgoingTxBatch{}
	_ EthereumSigned = &OutgoingLogicCall{}
)
