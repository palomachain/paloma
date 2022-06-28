#!/bin/bash
set -euo pipefail
set -x

if which gsed > /dev/null; then
  SED="$(which gsed)"
else
  SED="$(which sed)"
fi

if ! which jq > /dev/null; then
  echo 'command jq not found, please install jq'
  exit 1
fi

CHAIN_ID="paloma"

VALIDATOR_ACCOUNT_NAME="my_validator"
VALIDATOR_STAKE_AMOUNT="100000000ugrain"

if [[ -z ${PALOMA_CMD:-} ]]; then
  echo "building palomad binary"
  go build -o /tmp/palomad ./cmd/palomad 
  PALOMA="/tmp/palomad"
else
  PALOMA="go run ./cmd/palomad"
fi

$PALOMA init my_validator --chain-id "$CHAIN_ID"

pushd ~/.paloma/config/
$SED -i 's/keyring-backend = ".*"/keyring-backend = "test"/' client.toml
$SED -i 's/minimum-gas-prices = ".*"/minimum-gas-prices = "0.001ugrain"/' app.toml
$SED -i 's/laddr = ".*:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/' config.toml
jq ".chain_id = \"${CHAIN_ID}\"" genesis.json > temp.json && mv temp.json genesis.json
popd

if [[ -z "${ACCOUNT_MNEMONIC:-}" ]]; then
  $PALOMA keys add "$VALIDATOR_ACCOUNT_NAME"
else
  echo "$ACCOUNT_MNEMONIC" | $PALOMA keys add "$VALIDATOR_ACCOUNT_NAME" --recover
fi

validator_address="$($PALOMA keys show "$VALIDATOR_ACCOUNT_NAME" -a)"

$PALOMA add-genesis-account "$validator_address" 100000000000000ugrain
$PALOMA gentx "$VALIDATOR_ACCOUNT_NAME" "$VALIDATOR_STAKE_AMOUNT" --chain-id "$CHAIN_ID"
$PALOMA collect-gentxs
