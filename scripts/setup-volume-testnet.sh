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

if [[ -z "${MNEMONIC:-}" ]]; then
  echo 'MNEMONIC required'
  exit 1
fi

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
popd

GR=000000ugrain
MGR="000000${GR}"

INIT_AMOUNT="5${MGR}"
INIT_VALIDATION_AMOUNT="1${MGR}"
GENESIS_AMOUNT="1${MGR}"
FAUCET_AMOUNT="3${MGR}"

name="chase"
echo "$MNEMONIC" | palomad keys add "$name" --recover
address="$(palomad keys show "$name" -a)"

palomad add-genesis-account "$address" "$INIT_AMOUNT"
palomad gentx "$name" "$INIT_VALIDATION_AMOUNT" --chain-id "$CHAIN_ID"

init() {
  name="$1"
  address="$2"
  amount="${3:-"$GENESIS_AMOUNT"}"

  palomad add-genesis-account "$address" "$amount"
}

init faucet paloma167rf0jmkkkqp9d4awa8rxw908muavqgghtw6tn "$FAUCET_AMOUNT"
init jason paloma1mre80u0mmsdpf3l2shre9g4sh7kp9lxu5gtlql
init matija paloma1auhjn3rd6edsv90h72xuqnh7xwhlcayjselctn
init taariq paloma1k4hfe8cqdzy6j0t7sppujhrplhp8k5tglf8v46
init vera paloma132nvgqw2l2jfgwzvtxyt8fswhfwtw9dz76zxcy

palomad collect-gentxs
