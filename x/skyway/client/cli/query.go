package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/skyway/types"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	FlagOrder     = "order"
	FlagClaimType = "claim-type"
	FlagNonce     = "nonce"
	FlagEthHeight = "eth-height"
)

// GetQueryCmd bundles all the query subcmds together so they appear under `skyway query` or `skyway q`
func GetQueryCmd() *cobra.Command {
	// nolint: exhaustruct
	skywayQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the skyway module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	skywayQueryCmd.AddCommand([]*cobra.Command{
		CmdGetErc20ToDenoms(),
		CmdGetPendingOutgoingTXBatchRequest(),
		CmdGetOutgoingTXBatchRequest(),
		CmdGetPendingSendToRemote(),
		CmdGetAttestations(),
		CmdGetLastObservedEthBlock(),
		CmdGetLastObservedEthNonce(),
		CmdGetLastObservedEthNoncesByValAddress(),
		CmdGetQueryParams(),
		CmdGetQueryBridgeTax(),
		CmdGetQueryBridgeTransferLimits(),
		CmdGetLightNodeSaleContracts(),
	}...)

	return skywayQueryCmd
}

func CmdGetErc20ToDenoms() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "erc20-to-denoms",
		Short: "Query mapping of erc20 token addresses to denoms",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryErc20ToDenoms{}

			res, err := queryClient.GetErc20ToDenoms(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdGetPendingOutgoingTXBatchRequest fetches the batch to be confirmed next by the given validator, if any exists
func CmdGetPendingOutgoingTXBatchRequest() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "pending-batch-request [bech32 orchestrator address]",
		Short: "Get the latest outgoing TX batch request which has not been signed by a particular orchestrator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryLastPendingBatchRequestByAddrRequest{
				Address: args[0],
			}

			res, err := queryClient.LastPendingBatchRequestByAddr(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdGetOutgoingTXBatchRequest fetches all outgoing tx batches
func CmdGetOutgoingTXBatchRequest() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "outgoing-tx-batches [chain-reference-id] [orchestrator-address]",
		Short: "Get all outgoing TX batches",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryOutgoingTxBatchesRequest{
				ChainReferenceId: args[0],
				Assignee:         args[1],
			}

			res, err := queryClient.OutgoingTxBatches(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdGetPendingSendToRemote fetches all pending Sends to Ethereum made by the given address
func CmdGetPendingSendToRemote() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "pending-txs [from-address]",
		Short: "Query pending outgoing transactions from address.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryPendingSendToRemote{
				SenderAddress: args[0],
			}

			res, err := queryClient.GetPendingSendToRemote(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdGetAttestations fetches the most recently created Attestations in the store (only the most recent 1000 are available)
// up to an optional limit
func CmdGetAttestations() *cobra.Command {
	short := "Query current and historical skyway attestations (only the most recent 1000 are stored)"
	long := short + "\n\nOptionally provide a limit to reduce the number of attestations returned"

	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "attestations [chain-reference-id] [optional limit]",
		Args:  cobra.MinimumNArgs(1),
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			chainReferenceID := args[0]
			var limit uint64
			// Limit is 0 or whatever the user put in
			if len(args) < 2 || args[1] == "" {
				limit = 0
			} else {
				limit, err = strconv.ParseUint(args[1], 10, 64)
				if err != nil {
					return err
				}
			}
			orderBy, err := cmd.Flags().GetString(FlagOrder)
			if err != nil {
				return err
			}
			claimType, err := cmd.Flags().GetString(FlagClaimType)
			if err != nil {
				return err
			}
			nonce, err := cmd.Flags().GetUint64(FlagNonce)
			if err != nil {
				return err
			}
			height, err := cmd.Flags().GetUint64(FlagEthHeight)
			if err != nil {
				return err
			}

			req := &types.QueryAttestationsRequest{
				Limit:            limit,
				OrderBy:          orderBy,
				ClaimType:        claimType,
				Nonce:            nonce,
				Height:           height,
				ChainReferenceId: chainReferenceID,
			}
			res, err := queryClient.GetAttestations(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	// Global flags
	flags.AddQueryFlagsToCmd(cmd)
	// Local flags
	cmd.Flags().String(FlagOrder, "asc", "order attestations by eth block height: set to 'desc' for reverse ordering")
	cmd.Flags().String(FlagClaimType, "", "which types of claims to filter, empty for all or one of: CLAIM_TYPE_SEND_TO_PALOMA, CLAIM_TYPE_BATCH_SEND_TO_ETH")
	cmd.Flags().Uint64(FlagNonce, 0, "the exact nonce to find, 0 for any")
	cmd.Flags().Uint64(FlagEthHeight, 0, "the exact ethereum block height an event happened at, 0 for any")

	return cmd
}

// CmdGetLastObservedEthBlock fetches the Ethereum block height for the most recent "observed" Attestation, indicating
// the state of Cosmos consensus on the submitted Ethereum events
// nolint: dupl
func CmdGetLastObservedEthBlock() *cobra.Command {
	short := "Query the last observed block height on the remote chain"
	long := short + "\n\nThis value is expected to lag behind actual remote block height significantly due to 1. Remote Chain Finality and 2. Consensus mirroring the state."

	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "last-observed-block [chain-reference-id]",
		Short: short,
		Long:  long,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryLastObservedSkywayBlockRequest{
				ChainReferenceId: args[0],
			}
			res, err := queryClient.LastObservedSkywayBlock(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdGetLastObservedEthNonce fetches the Ethereum event nonce for the most recent "observed" Attestation, indicating
// // the state of Cosmos consensus on the submitted Ethereum events
// nolint: dupl
func CmdGetLastObservedEthNonce() *cobra.Command {
	short := "Query the last observed event nonce on the remote chain."
	long := short + "\n\nThis value is expected to lag behind actual remote block height significantly due to 1. Remote Chain Finality and 2. Consensus mirroring the state."

	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "last-observed-nonce [chain-reference-id]",
		Short: short,
		Long:  long,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryLastObservedSkywayNonceRequest{
				ChainReferenceId: args[0],
			}
			res, err := queryClient.LastObservedSkywayNonce(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdGetLastObservedEthNoncesByValAddress() *cobra.Command {
	short := "Query the last observed event nonce on the remote chain."
	long := short + "\n\nThis value is expected to lag behind actual remote block height significantly due to 1. Remote Chain Finality and 2. Consensus mirroring the state."

	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "last-observed-nonce-by-val [chain-reference-id] [val-addr]",
		Short: short,
		Long:  long,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryLastObservedSkywayNonceByAddrRequest{
				Address:          args[1],
				ChainReferenceId: args[0],
			}
			res, err := queryClient.LastObservedSkywayNonceByAddr(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdGetQueryParams fetches the current Skyway module params
func CmdGetQueryParams() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "",
		Args:  cobra.NoArgs,
		Short: "Query skyway params",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdGetQueryBridgeTax fetches the current Skyway module bridge tax settings
func CmdGetQueryBridgeTax() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridge-taxes",
		Short: "Query bridge tax settings for all tokens",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &emptypb.Empty{}

			res, err := queryClient.GetBridgeTaxes(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdGetQueryBridgeTransferLimits fetches transfer limits for all tokens
func CmdGetQueryBridgeTransferLimits() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridge-transfer-limits",
		Short: "Query bridge transfer limits for all tokens",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &emptypb.Empty{}

			res, err := queryClient.GetBridgeTransferLimits(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetLightNodeSaleContracts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "light-node-sale-contracts",
		Short: "Query all light node sale contracts details",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &emptypb.Empty{}

			res, err := queryClient.GetLightNodeSaleContracts(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
