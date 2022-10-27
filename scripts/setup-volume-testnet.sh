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

palomad init birdlady --chain-id "$CHAIN_ID"

pushd ~/.paloma/config/
sed -i 's/^keyring-backend = ".*"/keyring-backend = "test"/' client.toml
sed -i 's/^minimum-gas-prices = ".*"/minimum-gas-prices = "0.001ugrain"/' app.toml
sed -i 's/^laddr = ".*:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/' config.toml
jq-i ".chain_id = \"${CHAIN_ID}\"" genesis.json
popd

GR=000000ugrain
KGR="000${GR}"
MGR="000000${GR}"

INIT_AMOUNT="5${MGR}"
INIT_VALIDATION_AMOUNT="10${KGR}"
GENESIS_AMOUNT="1${MGR}"
FAUCET_AMOUNT="3${MGR}"

name="splat"
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
init chase paloma1nty4gn8k2nrewy26fm62v03322fxgpq0hxssn6
init jason paloma1mre80u0mmsdpf3l2shre9g4sh7kp9lxu5gtlql
init taariq paloma1k4hfe8cqdzy6j0t7sppujhrplhp8k5tglf8v46

init  1 paloma1hj4lxqp05ntjl6ezu3qn9eyedp5p4fpdwu8gxj "$INIT_VALIDATION_AMOUNT"
init  2 paloma1d3v3jh6l2r23y9kgzdrahx0ev8ez0g8qapsfxs "$INIT_VALIDATION_AMOUNT"
init  3 paloma16jhsvx4zrukkjqd9akfawx2tzduyx4uxgtjj42 "$INIT_VALIDATION_AMOUNT"
init  4 paloma15gvyk43x406v7kcd4rff5qfutqmcnpj3p4ea9g "$INIT_VALIDATION_AMOUNT"
init  5 paloma1z9fgzh7mzqgu33pdkxw0dqmqgm9l8exj4nggcp "$INIT_VALIDATION_AMOUNT"
init  6 paloma1ljg6ed0pzc3xpqtareyfp6h4fpngs7nw0nnu2j "$INIT_VALIDATION_AMOUNT"
init  7 paloma13uslh0y22ffnndyr3x30wqd8a6peqh255hkzrz "$INIT_VALIDATION_AMOUNT"
init  8 paloma1qf7np0rp3qutvn08vc6qz0y7ffc02cncpam3a7 "$INIT_VALIDATION_AMOUNT"
init  9 paloma10y227j9d09pckexy32v2gckerj9a0kcepc7zsh "$INIT_VALIDATION_AMOUNT"
init 10	paloma1swa5kcf9cl5dx2ypx0c5r9e5qdfnzp9wq59z2l "$INIT_VALIDATION_AMOUNT"
init 12	paloma1espezfvhuagml6g0flw8jj6ajs2wud4cy7wgdu "$INIT_VALIDATION_AMOUNT"
init 13	paloma1tdw23fpnxh2uk3djtteh7eaydymrfgnak3paq3 "$INIT_VALIDATION_AMOUNT"
init 14	paloma1pmx8m7glpw4pckwzrfflqp4427ynl32wam45z4 "$INIT_VALIDATION_AMOUNT"
init 15	paloma1ylmkqmwct392ytu52tncpeyz4dskklx56cd3rf "$INIT_VALIDATION_AMOUNT"
init 16	paloma16lez38lgsgu34ka2gq8yee8a862zpgs4rt52xs "$INIT_VALIDATION_AMOUNT"
init 17	paloma1am3k7czusdcewv55nhaugu2drn38af9449yxzj "$INIT_VALIDATION_AMOUNT"
init 18	paloma1pdd0nmuj0xfwhfyt7h3wkx9zgjvs3hzlehakd5 "$INIT_VALIDATION_AMOUNT"
init 19	paloma1kludne80z0tcq9t7j9fqa630fechsjhxhpafac "$INIT_VALIDATION_AMOUNT"
init 20	paloma14jux7zw5qd9gdrapypgndwysrmn29gx2nzsf70 "$INIT_VALIDATION_AMOUNT"

palomad collect-gentxs
