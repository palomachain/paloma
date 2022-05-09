#!/usr/bin/env bash

set -ex

CHAIN_ID="paloma"

VALIDATOR_ACCOUNT_NAME="my_validator"
VALIDATOR_STAKE_AMOUNT="1000000dove"

PALOMA="${PALOMA_CMD:-go run ./cmd/palomad}"

$PALOMA init my_validator --chain-id $CHAIN_ID

pushd ~/.paloma/config/
sed -i 's/keyring-backend = ".*"/keyring-backend = "test"/' client.toml
sed -i 's/minimum-gas-prices = ".*"/minimum-gas-prices = "0.001dove"/' app.toml
sed -i 's/laddr = ".*:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/' config.toml
jq ".chain_id = \"${CHAIN_ID}\"" genesis.json > temp.json && mv temp.json genesis.json
jq 'walk(if type == "string" and .. == "stake" then "dove" else . end)' genesis.json > temp.json && mv temp.json genesis.json
popd

$PALOMA keys add $VALIDATOR_ACCOUNT_NAME

validator_address=$($PALOMA keys show $VALIDATOR_ACCOUNT_NAME -a)

$PALOMA add-genesis-account $validator_address 100000000000000dove
$PALOMA gentx $VALIDATOR_ACCOUNT_NAME $VALIDATOR_STAKE_AMOUNT --chain-id $CHAIN_ID
$PALOMA collect-gentxs
