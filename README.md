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

## Testnet Setup Instructions

To get the latest `palomad` binary:

```shell
wget -O - https://github.com/palomachain/paloma/releases/download/v0.2.5-prealpha/paloma_0.2.5-prealpha_Linux_x86_64.tar.gz | \
sudo tar -C /usr/local/bin -xvzf - palomad
sudo chmod +x /usr/local/bin/palomad
# Required until we figure out cgo
sudo wget -P /usr/lib https://github.com/CosmWasm/wasmvm/raw/main/api/libwasmvm.x86_64.so
```

### Connecting to an existing testnet.

Download and install the latest release of palomad.

Initialize our configuration. This will populate a `~/.paloma/` directory.
```shell
MONIKER="$(hostname)"
palomad init "$MONIKER"
```

Copy the configs of the testnet we wish to connect to

```shell
wget -O ~/.paloma/config/genesis.json https://raw.githubusercontent.com/palomachain/testnet/master/paloma-testnet-6/genesis.json
wget -O ~/.paloma/config/addrbook.json https://raw.githubusercontent.com/palomachain/testnet/master/paloma-testnet-6/addrbook.json
```

Next you can generate a new set of keys to the new machine, or reuse an existing key.
```shell
VALIDATOR=<choose a name>
palomad keys add "$VALIDATOR"

# Or if you have a mnemonic already, we can recover the keys with:
palomad keys add "$VALIDATOR" --recover
```

Head over to https://faucet.palomaswap.com/ and get some funds!

We can verify the new funds have been deposited.
```shell
palomad query bank balances --node tcp://testnet.palomaswap.com:26657 "$ADDRESS"
```

And start the node!
```shell
palomad start
```

If desired we can stake our funds and create a validator.
```shell
MONIKER="$(hostname)"
VALIDATOR="$(palomad keys list --list-names | head -n1)"
STAKE_AMOUNT=1000000ugrain
PUBKEY="$(palomad tendermint show-validator)"
palomad tx staking create-validator \
      --fees=1000000ugrain \
      --from="$VALIDATOR" \
      --amount="$STAKE_AMOUNT" \
      --pubkey="$PUBKEY" \
      --moniker="$MONIKER" \
      --website="https://www.example.com" \
      --details="<enter a description>" \
      --chain-id=paloma-testnet-6 \
      --commission-rate="0.1" \
      --commission-max-rate="0.2" \
      --commission-max-change-rate="0.05" \
      --min-self-delegation="100" \
      --yes \
      --broadcast-mode=block
```

You may receive an error `account sequence mismatch`, you will need to wait until your local paloma
catches up with the rest of the chain.

#### Running with `systemd`

First configure the service:

```shell
cat << EOT > /etc/systemd/system/palomad.service
[Unit]
Description=Paloma Blockchain
After=network.target
ConditionPathExists=/usr/local/bin/palomad

[Service]
Type=simple
LimitNOFILE=65535
Restart=always
RestartSec=5
WorkingDirectory=~
ExecStartPre=
ExecStart=/usr/local/bin/palomad start
ExecReload=

[Install]
WantedBy=multi-user.target
EOT
```

Then reload systemd configurations and start the service!

```shell
service palomad start

# Check that it's running successfully:
service palomad status
# Or watch the logs:
journalctl -u palomad.service -f
```

### Uploading a local contract

```shell
CONTRACT=<contract.wasm>
VALIDATOR="$(palomad keys list --list-names | head -n1)"
palomad tx wasm store "$CONTRACT" --from "$VALIDATOR" --broadcast-mode block -y --gas auto --fees 3000000000ugrain
```

## Setting up a new private testnet

Download and install the latest release of palomad.

Install `jq`, used by the setup script.

```shell
apt install jq
```

Set up the chain validator.

```shell
CHAIN_ID=paloma-testnet-6 \
MNEMONIC="$(cat secret.mn)" \
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/palomachain/paloma/master/scripts/setup-volume-testnet.sh)"
```

We should now see an error free execution steadily increasing chain depth.

```shell
palomad start
```

Others wishing to connect to the new testnet will need your `.paloma/config/genesis.json` file,
as well as the main peer designation. We can get the main peer designation with `jq`:

```shell
jq -r '.body.memo' ~/.paloma/config/gentx/gentx-*.json
```
