#!/bin/bash
set -euo pipefail
set -x

if ! which jq > /dev/null; then
  echo 'command jq not found, please install jq'
  exit 1
fi

if [[ -z "${CHAIN_ID:-}" ]]; then
  echo 'CHAIN_ID required'
  exit 1
fi

VALIDATOR_STAKE_AMOUNT=100000000000000ugrain
GENESIS_AMOUNT=2500000000000000ugrain
FAUCET_AMOUNT=500000000000000ugrain

jq-i() {
  edit="$1"
  f="$2"
  jq "$edit" "$f" > "${f}.tmp"
  mv "${f}.tmp" "$f"
}

palomad init my_validator --chain-id "$CHAIN_ID"

pushd ~/.paloma/config/
sed -i 's/^keyring-backend = ".*"/keyring-backend = "test"/' client.toml
sed -i 's/^minimum-gas-prices = ".*"/minimum-gas-prices = "0.001ugrain"/' app.toml
sed -i 's/^laddr = ".*:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/' config.toml
jq-i ".chain_id = \"${CHAIN_ID}\"" genesis.json
jq-i 'walk(if type == "string" and .. == "stake" then "grain" else . end)' genesis.json
popd

name="chase"
echo "$MNEMONIC" | palomad keys add "$name" --recover
address="$(palomad keys show "$name" -a)"

palomad add-genesis-account "$address" "$GENESIS_AMOUNT"
palomad gentx "$name" "$VALIDATOR_STAKE_AMOUNT" --chain-id "$CHAIN_ID"

init() {
  name="$1"
  address="$2"
  amount="${3:-"$GENESIS_AMOUNT"}"

  palomad add-genesis-account "$address" "$amount"
}

init faucet paloma167rf0jmkkkqp9d4awa8rxw908muavqgghtw6tn "$FAUCET_AMOUNT"
init jason paloma1mre80u0mmsdpf3l2shre9g4sh7kp9lxu5gtlql
init taariq paloma1k4hfe8cqdzy6j0t7sppujhrplhp8k5tglf8v46
init vera paloma1nvhryyg63e8vw9t5sx9ht8l77jzgjvrq6jcdnp

palomad collect-gentxs
