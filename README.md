![Logo!](assets/paloma.png)

[![Continuous integration](https://github.com/palomachain/paloma/actions/workflows/ci-test.yml/badge.svg?branch=master)](https://github.com/palomachain/paloma/actions/workflows/ci-test.yml)
[![Project Status: WIP – Initial development is in progress, but there has not yet been a stable, usable release suitable for the public.](https://img.shields.io/badge/repo%20status-WIP-yellow.svg?style=flat-square)](https://www.repostatus.org/#wip)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/palomachain/paloma?logo=paloma)
[![License: Apache-2.0](https://img.shields.io/github/license/umee-network/umee.svg?style=flat-square)](https://github.com/palomachain/paloma/blob/main/LICENSE)
![Lines of code](https://img.shields.io/tokei/lines/github/palomachain/paloma)

> A Golang implementation of Paloma Chain, a decentralized, automation network for smart contracts
> deployed in the Cosmos, EVM, Solana, and Polkadot networks.

For Crosschain Software engineers that want simultaneous control of mulitiple smart contracts, on any blockchain, Paloma is decentralized and consensus-driven message delivery, fast state awareness, low cost state computation, and powerful attestation system that enables scaleable, crosschain, smart contract execution with any data source.


## Table of Contents

- [Talk To Us](#talk-to-us)
- [Releases](#releases)
- [Active Networks](#active-networks)
- [Public](#public)
- [Private](#private)
- [Install](#install)
- [Contributing](CONTRIBUTING.md)

## Talk to us

We have active, helpful communities on Twitter and Telegram.

* [Twitter](https://twitter.com/paloma_chain)
* [Telegram](https://t.me/palomachain)

## Releases

See [Release procedure](CONTRIBUTING.md#release-procedure) for more information about the release model.

## Active Networks

### Mainnet

N/A

## Install

To download the `palomad` binary:

```shell
sudo wget -O - https://github.com/palomachain/paloma/releases/download/v0.0.1-alphawasmd4/paloma_0.0.1-alphawasmd4_Linux_x86_64.tar.gz | \
sudo tar -C /usr/local/bin -xvzf - palomad
sudo chmod +x /usr/local/bin/palomad
# Required until we figure out cgo
sudo wget -P /usr/lib https://github.com/CosmWasm/wasmvm/raw/main/api/libwasmvm.x86_64.so
```

## Setting up a new chain from scratch.

### Initializing a new chain

Download and install the latest release of palomad.

Install `jq`, used by the setup script.

```shell
apt install jq
```

Set up the chain validator.

```shell
PALOMA_CMD=/PATH/TO/palomad \
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/palomachain/paloma/master/scripts/setup-chain-validator.sh)"
```

We should now see an error free execution steadily increasing chain depth.

```shell
palomad start
```

### Connecting to an existing chain

Download and install the latest release of palomad.

Initialize our configuration. This will populate a `~/.paloma/` directory.
```shell
MONIKER="$(hostname)"
palomad init "$MONIKER"
```

Copy several files from the original server to the new server:
  - `.paloma/config/app.toml`
  - `.paloma/config/client.toml`
  - `.paloma/config/config.toml`
  - `.paloma/config/genesis.json`

Next you will need to add a new set of keys to the new machine.
```shell
VALIDATOR=<choose a name>
palomad keys add "$VALIDATOR"
```

Transfer yourself some funds. We'll need the address of our new validator and the address of the original validator.
You can run `palomad keys list` on their respective nodes. On the main node run:

```shell
palomad --fees 200dove tx bank send "$ORIGINAL_VALIDATOR_ADDRESS" "$VALIDATOR_ADDRESS" 100000000dove
# Verify that we see the funds:
palomad query bank balances "$VALIDATOR_ADDRESS"
```

Stake your funds and start `palomad`.
```shell
MAIN_VALIDATOR_NODE="tcp://157.245.76.119:26657"
CHAIN_ID="paloma"
DEFAULT_GAS_AMOUNT="10000000"
VALIDATOR_STAKE_AMOUNT=100000dove
PUBKEY="$(palomad tendermint show-validator)"
palomad tx staking create-validator \
      --fees 10000dove \
      --from="$VALIDATOR" \
      --amount="$VALIDATOR_STAKE_AMOUNT" \
      --pubkey="$PUBKEY" \
      --moniker="$MONIKER" \
      --chain-id="$CHAIN_ID" \
      --commission-rate="0.1" \
      --commission-max-rate="0.2" \
      --commission-max-change-rate="0.05" \
      --min-self-delegation="100" \
      --gas=$DEFAULT_GAS_AMOUNT \
      --node "$MAIN_VALIDATOR_NODE" \
      --yes \
      -b block
```

Run `jq -r '.body.memo' ~/.paloma/config/gentx/gentx-*.json` on the original validator for its peer designation,
and start our validator:

```shell
palomad start --p2p.persistent_peers "$MAIN_PEER_DESIGNATION"
```
