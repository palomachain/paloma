package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/palomachain/paloma/x/gravity/types"
	"github.com/spf13/cobra"
)

const (
	FlagOrder     = "order"
	FlagClaimType = "claim-type"
	FlagNonce     = "nonce"
	FlagEthHeight = "eth-height"
	FlagUseV1Key  = "use-v1-key"
)

// GetQueryCmd bundles all the query subcmds together so they appear under `gravity query` or `gravity q`
func GetQueryCmd() *cobra.Command {
	// nolint: exhaustruct
	gravityQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the gravity module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	gravityQueryCmd.AddCommand([]*cobra.Command{
		CmdGetErc20ToDenoms(),
		CmdGetPendingOutgoingTXBatchRequest(),
		CmdGetOutgoingTXBatchRequest(),
		CmdGetPendingSendToEth(),
		CmdGetAttestations(),
		CmdGetLastObservedEthBlock(),
		CmdGetLastObservedEthNonce(),
		GetCmdQueryParams(),
	}...)

	return gravityQueryCmd
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

// CmdGetPendingSendToEth fetches all pending Sends to Ethereum made by the given address
func CmdGetPendingSendToEth() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "pending-send-to-eth [address]",
		Short: "Query transactions waiting to go to Ethereum",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryPendingSendToEth{
				SenderAddress: args[0],
			}

			res, err := queryClient.GetPendingSendToEth(cmd.Context(), req)
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
	short := "Query gravity current and historical attestations (only the most recent 1000 are stored)"
	long := short + "\n\n" + "Optionally provide a limit to reduce the number of attestations returned" + "\n" +
		"Note that when querying with --height less than 1282013 '--use-v1-key' must be provided to locate the attestations"

	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "attestations [optional limit]",
		Args:  cobra.MaximumNArgs(1),
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			var limit uint64
			// Limit is 0 or whatever the user put in
			if len(args) == 0 || args[0] == "" {
				limit = 0
			} else {
				limit, err = strconv.ParseUint(args[0], 10, 64)
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
			useV1Key, err := cmd.Flags().GetBool(FlagUseV1Key)
			if err != nil {
				return err
			}

			req := &types.QueryAttestationsRequest{
				Limit:     limit,
				OrderBy:   orderBy,
				ClaimType: claimType,
				Nonce:     nonce,
				Height:    height,
				UseV1Key:  useV1Key,
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
	cmd.Flags().Bool(FlagUseV1Key, false, "if querying with --height less than 1282013 this flag must be provided to locate the attestations")

	return cmd
}

// CmdGetLastObservedEthBlock fetches the Ethereum block height for the most recent "observed" Attestation, indicating
// the state of Cosmos consensus on the submitted Ethereum events
// nolint: dupl
func CmdGetLastObservedEthBlock() *cobra.Command {
	short := "Query the last observed Ethereum block height"
	long := short + "\n\n" +
		"This value is expected to lag the actual Ethereum block height significantly due to 1. Ethereum Finality and 2. Consensus mirroring the state on Ethereum" + "\n" +
		"Note that when querying with --height less than 1282013 '--use-v1-key' must be provided to locate the value"

	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "last-observed-eth-block",
		Short: short,
		Long:  long,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			useV1Key, err := cmd.Flags().GetBool(FlagUseV1Key)
			if err != nil {
				return err
			}

			req := &types.QueryLastObservedEthBlockRequest{
				UseV1Key: useV1Key,
			}
			res, err := queryClient.GetLastObservedEthBlock(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Bool(FlagUseV1Key, false, "if querying with --height less than 1282013 this flag must be provided to locate the Last Observed Ethereum Height")
	return cmd
}

// CmdGetLastObservedEthNonce fetches the Ethereum event nonce for the most recent "observed" Attestation, indicating
// // the state of Cosmos consensus on the submitted Ethereum events
// nolint: dupl
func CmdGetLastObservedEthNonce() *cobra.Command {
	short := "Query the last observed Ethereum event nonce"
	long := short + "\n\n" +
		"This this is likely to lag the last executed event a little due to 1. Ethereum Finality and 2. Consensus mirroring the Ethereum state" + "\n" +
		"Note that when querying with --height less than 1282013 '--use-v1-key' must be provided to locate the value"

	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "last-observed-eth-nonce",
		Short: short,
		Long:  long,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			useV1Key, err := cmd.Flags().GetBool(FlagUseV1Key)
			if err != nil {
				return err
			}

			req := &types.QueryLastObservedEthNonceRequest{
				UseV1Key: useV1Key,
			}
			res, err := queryClient.GetLastObservedEthNonce(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Bool(FlagUseV1Key, false, "if querying with --height less than 1282013 this must be set to true to locate the Last Observed Ethereum Event Nonce")
	return cmd
}

// GetCmdQueryParams fetches the current Gravity module params
func GetCmdQueryParams() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query gravity params",
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
