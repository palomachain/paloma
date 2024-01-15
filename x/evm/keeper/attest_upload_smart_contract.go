package keeper

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/evm/types"
)

type attestionParameters struct {
	rawEvidence      any
	msg              *types.Message
	chainReferenceID string
	msgID            uint64
}

type uploadSmartContractAttester struct {
	attestionParameters
	action *types.UploadSmartContract
	logger liblog.Logr
	k      *Keeper
}

func newUploadSmartContractAttester(k *Keeper, l liblog.Logr, p attestionParameters) *uploadSmartContractAttester {
	return &uploadSmartContractAttester{
		attestionParameters: p,
		logger:              l,
		k:                   k,
	}
}

func (a *uploadSmartContractAttester) Execute(ctx sdk.Context) error {
	a.logger = a.logger.WithFields("action-msg", "Message_UploadSmartContract")
	a.logger.Debug("Processing upload smart contract message attestation.")

	a.action = a.msg.Action.(*types.Message_UploadSmartContract).UploadSmartContract

	switch evidence := a.rawEvidence.(type) {
	case *types.TxExecutedProof:
		return a.attest(ctx, evidence)
	case *types.SmartContractExecutionErrorProof:
		a.logger.Debug("smart contract execution error proof", "smart-contract-error", evidence.GetErrorMessage())
		keeperutil.EmitEvent(a.k, ctx, types.SmartContractExecutionFailedKey,
			types.SmartContractExecutionFailedMessageID.With(fmt.Sprintf("%d", a.msgID)),
			types.SmartContractExecutionFailedChainReferenceID.With(a.chainReferenceID),
			types.SmartContractExecutionFailedError.With(evidence.GetErrorMessage()),
			types.SmartContractExecutionMessageType.With(fmt.Sprintf("%T", a.action)),
		)
	default:
		a.logger.Error("unknown type %t when attesting", evidence)
		return ErrUnexpectedError.JoinErrorf("unknown type %t when attesting", evidence)
	}

	return nil
}

func (a *uploadSmartContractAttester) attest(ctx sdk.Context, evidence *types.TxExecutedProof) error {
	tx, err := attestTransactionIntegrity(ctx, a.k, evidence, a.action.VerifyAgainstTX)
	if err != nil {
		a.logger.WithError(err).Error("Failed to verify transaction integrity.")
		return err
	}

	smartContractID := a.action.GetId()
	deployment, _ := a.k.getSmartContractDeploymentByContractID(ctx, smartContractID, a.chainReferenceID)
	if deployment == nil {
		a.logger.WithError(err).WithFields("smart-contract-id", smartContractID).Error("Smart contract not found")
		return ErrCannotActiveSmartContractThatIsNotDeploying
	}

	if deployment.GetStatus() != types.SmartContractDeployment_IN_FLIGHT {
		a.logger.WithError(err).Error("deployment not in right state")
		return ErrCannotActiveSmartContractThatIsNotDeploying
	}

	// Update deployment
	ethMsg, err := core.TransactionToMessage(tx, ethtypes.NewLondonSigner(tx.ChainId()), big.NewInt(0))
	if err != nil {
		a.logger.WithError(err).Error("Failed to extract ethMsg")
		return err
	}
	newCompassAddr := crypto.CreateAddress(ethMsg.From, tx.Nonce())
	deployment.NewSmartContractAddress = newCompassAddr.Hex()
	deployment.Status = types.SmartContractDeployment_WAITING_FOR_ERC20_OWNERSHIP_TRANSFER
	if err := a.k.updateSmartContractDeployment(ctx, smartContractID, a.chainReferenceID, deployment); err != nil {
		a.logger.WithError(err).Error("Failed to update smart contract deployment")
		return err
	}

	records, err := a.k.Gravity.CastAllERC20ToDenoms(ctx)
	if err != nil {
		a.logger.WithError(err).Error("Failed to extract ERC20 records.")
		return err
	}

	if len(records) > 0 {
		return a.startTokenRelink(ctx, deployment, records, newCompassAddr, smartContractID)
	}

	// We don't have any tokens on the target chain. Set contract as active immediately.
	return a.k.SetSmartContractAsActive(ctx, smartContractID, a.chainReferenceID)
}

func (a *uploadSmartContractAttester) startTokenRelink(
	ctx sdk.Context,
	deployment *types.SmartContractDeployment,
	records []types.ERC20Record,
	newCompassAddr common.Address,
	smartContractID uint64,
) error {
	msgIDs := make([]uint64, 0, len(records))
	transfers := make([]types.SmartContractDeployment_ERC20Transfer, 0, len(records))
	erc20abi := `[{"inputs": [{"name": "_compass","type": "address"}],"name": "new_compass","outputs": [],"stateMutability": "nonpayable","type": "function"}]`

	for _, v := range records {
		if v.GetChainReferenceId() != a.chainReferenceID {
			continue
		}

		payload, err := func() ([]byte, error) {
			evm, err := abi.JSON(strings.NewReader(erc20abi))
			if err != nil {
				return nil, err
			}
			return evm.Pack("new_compass", newCompassAddr)
		}()
		if err != nil {
			return err
		}

		// SLCs are usually always authored by either a contract on Paloma, or
		// a specific validator. In this case, this is really a consensus operation
		// without a singular governing entity. For the sake the established
		// technological boundaries, we'll set the sender to the address of the
		// validator that attested this message.
		valAddr, err := sdk.ValAddressFromBech32(a.msg.GetAssignee())
		if err != nil {
			return fmt.Errorf("validator address from bech32: %w", err)
		}

		sender := sdk.AccAddress(valAddr.Bytes())

		modifiedPayload, err := injectSenderIntoPayload(make([]byte, 32), payload)
		if err != nil {
			return fmt.Errorf("inject zero padding to payload: %w", err)
		}

		ci, err := a.k.GetChainInfo(ctx, a.chainReferenceID)
		if err != nil {
			return fmt.Errorf("get chain info: %w", err)
		}

		msgID, err := a.k.AddSmartContractExecutionToConsensus(
			ctx,
			a.chainReferenceID,
			string(ci.GetSmartContractUniqueID()),
			&types.SubmitLogicCall{
				HexContractAddress: v.GetErc20(),
				Abi:                common.FromHex(""),
				Payload:            modifiedPayload,
				Deadline:           ctx.BlockTime().Add(10 * time.Minute).Unix(),
				SenderAddress:      sender,
				ExecutionRequirements: types.SubmitLogicCall_ExecutionRequirements{
					EnforceMEVRelay: false,
				},
			},
		)
		if err != nil {
			return fmt.Errorf("execute job: %w", err)
		}

		msgIDs = append(msgIDs, msgID)
		transfers = append(transfers, types.SmartContractDeployment_ERC20Transfer{
			Denom:  v.GetDenom(),
			Erc20:  v.GetErc20(),
			MsgID:  msgID,
			Status: types.SmartContractDeployment_ERC20Transfer_PENDING,
		})
	}

	deployment.Erc20Transfers = transfers
	if err := a.k.updateSmartContractDeployment(ctx, smartContractID, a.chainReferenceID, deployment); err != nil {
		a.logger.WithError(err).Error("Failed to update smart contract deployment")
		return err
	}

	a.k.deploymentCache.Add(ctx, a.chainReferenceID, smartContractID, msgIDs...)
	a.logger.Debug("attestation successful")
	return nil
}
