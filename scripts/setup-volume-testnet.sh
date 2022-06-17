#!/bin/bash
set -euo pipefail
set -x

if ! which jq > /dev/null; then
  echo 'command jq not found, please install jq'
  exit 1
fi

CHAIN_ID="paloma"
VALIDATOR_STAKE_AMOUNT=100000000grain

palomad init my_validator --chain-id "$CHAIN_ID"

pushd ~/.paloma/config/
sed -i 's/^keyring-backend = ".*"/keyring-backend = "test"/' client.toml
sed -i 's/^minimum-gas-prices = ".*"/minimum-gas-prices = "0.001grain"/' app.toml
sed -i 's/^laddr = ".*:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/' config.toml
jq ".chain_id = \"${CHAIN_ID}\"" genesis.json > genesis.json.tmp && mv genesis.json.tmp genesis.json
jq 'walk(if type == "string" and .. == "stake" then "grain" else . end)' genesis.json > genesis.json.tmp && mv genesis.json.tmp genesis.json
popd

name="chase"
echo "$MNEMONIC" | palomad keys add "$name" --recover
address="$(palomad keys show "$name" -a)"

palomad add-genesis-account "$address" 250000000grain
palomad gentx "$name" "$VALIDATOR_STAKE_AMOUNT" --chain-id "$CHAIN_ID"

init() {
  name="$1"
  address="$2"
  amount="${3:-250000000grain}"

  palomad add-genesis-account "$address" "$amount"
}

init faucet paloma167rf0jmkkkqp9d4awa8rxw908muavqgghtw6tn 500000000grain
init jason paloma1mre80u0mmsdpf3l2shre9g4sh7kp9lxu5gtlql
init taariq paloma1k4hfe8cqdzy6j0t7sppujhrplhp8k5tglf8v46
init vera paloma1nvhryyg63e8vw9t5sx9ht8l77jzgjvrq6jcdnp

palomad collect-gentxs
