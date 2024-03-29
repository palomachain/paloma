<div align="center">
    <img alt="Paloma" src="https://github.com/palomachain/paloma/blob/master/assets/Paloma_black.png" />
</div>

> A Golang implementation of Paloma Chain, a decentralized, automation network for smart contracts
> deployed in the Cosmos, EVM, Solana, and Polkadot networks.

<div align="center">
  <a href="https://github.com/palomachain/paloma/blob/master/LICENSE">
    <img alt="License: Apache-2.0" src="https://img.shields.io/github/license/palomachain/paloma.svg" />
  </a>
  <a href="https://www.repostatus.org/#wip">
    <img alt="Project Status: WIP – Initial development is in progress, but there has not yet been a stable, usable release suitable for the public." src="https://img.shields.io/badge/repo%20status-WIP-yellow.svg?style=flat-square" />
  </a>
  <img alt="Go Version" src="https://img.shields.io/github/go-mod/go-version/palomachain/paloma?logo=paloma" />
  <a href="https://pkg.go.dev/github.com/palomachain/paloma">
    <img src="https://pkg.go.dev/badge/github.com/palomachain/paloma.svg" alt="Go Reference">
  </a>
  <a href="https://goreportcard.com/report/github.com/palomachain/paloma">
    <img alt="Go report card" src="https://goreportcard.com/badge/github.com/palomachain/paloma" />
  </a>
</div>
<div align="center">
  <a href="https://github.com/palomachain/paloma/actions/workflows/ci-test.yml">
    <img alt="Code Coverage" src="https://github.com/palomachain/paloma/actions/workflows/ci-test.yml/badge.svg" />
  </a>
  <a href="https://github.com/palomachain/paloma/actions/workflows/release.yml">
    <img alt="Code Coverage" src="https://github.com/palomachain/paloma/actions/workflows/release.yml/badge.svg" />
  </a>
  <img alt="Lines of code" src="https://img.shields.io/tokei/lines/github/palomachain/paloma" />
</div>

Paloma is the fastest, secure crosschain communications blockchain. For Crosschain Software engineers who want simultaneous control of multiple, cross-chain-deployed, smart contracts, Paloma is decentralized and consensus-driven message delivery. Paloma is fast state awareness, low cost state computation, and a powerful attestation system. Polama blockchain enables scalable, crosschain, smart contract execution with any data source.


## Table of Contents

- [Talk To Us](#talk-to-us)
- [International Community](#international-community)
- [Releases](#releases)
- [Active Networks](#active-networks)
- [Join a network](#join-an-active-network)
- [Contributing](CONTRIBUTING.md)

## Talk to us

We have active, helpful communities on Twitter and Telegram.

* [Twitter](https://twitter.com/paloma_chain)
* [Telegram](https://t.me/palomachain)
* [Discord](https://discord.gg/HtUvgxvh5N)
* [Forum](https://forum.palomachain.com/)

## International Community

- [Bengali README](docs/Welcome-Bengali.md)
- [Chinese README](docs/Welcome-Chinese.md)
- [Indonesian README](docs/Welcome-Indonesian.md)
- [Japanese README](docs/Welcome-Japanese.md)
- [Persian README](docs/Welcome-Persian.md)
- [Polish README](docs/Welcome-Polish.md)
- [Portuguese README](docs/Welcome-Portuguese.md)
- [Romanian README](docs/Welcome-Romanian.md)
- [Russian README](docs/Welcome-Russian.md)
- [Spanish README](docs/Welcome-Spanish.md)
- [Thai README](docs/Welcome-Thai.md)
- [Turkish README](docs/Welcome-Turkish.md)
- [Ukrainian README](docs/Welcome-Ukrainian.md)
- [Vietnamese README](docs/Welcome-Vietnamese.md)

## Releases

See [Release procedure](CONTRIBUTING.md#release-procedure) for more information about the release model.

## Active Networks
* Testnet `paloma-testnet-15` (January 20, 2023)
* Mainnet `messenger` (February 2, 2023)

## Join an active Network

> #### Note:
> Some have seen errors with GLIBC version differences with the downloaded binaries. This is caused by a difference in the libraries of the host that built the binary and the host running the binary.
>
> If you experience these errors, please pull down the code and build it, rather than downloading the prebuilt binary


### Install the correct version of libwasm
The current required version of libwasm is `1.5.2`. If you're upgrading from a prior version it is recommended to remove the cache to avoid errors. If you're already have `palomad` running, you will need to stop it before doing these steps.

```
wget https://github.com/CosmWasm/wasmvm/releases/download/v1.5.2/libwasmvm.x86_64.so
sudo mv libwasmvm.x86_64.so /usr/lib/

rm -r ~/.paloma/data/wasm/cache
```

### To get the latest prebuilt `palomad` binary:

```shell
wget -O - https://github.com/palomachain/paloma/releases/download/v1.13.0/paloma_Linux_x86_64.tar.gz  | \
  sudo tar -C /usr/local/bin -xvzf - palomad
sudo chmod +x /usr/local/bin/palomad
```

### To build palomad using latest release

```shell
git clone https://github.com/palomachain/paloma.git
cd paloma
git checkout v1.13.0
make build
sudo mv ./build/palomad /usr/local/bin/palomad
```

If you're upgrading to the most recent version, you will need to stop `palomad` before copying the new binary into place.

### Connecting to an existing network.

Download and install the latest release of palomad.

Initialize our configuration. This will populate a `~/.paloma/` directory.
```shell
MONIKER="$(hostname)"
palomad init "$MONIKER"

#for testnet
CHAIN_ID="paloma-testnet-15" 
#for mainnet
CHAIN_ID="messenger" 

```

Copy the configs of the network we wish to connect to

Testnet:
```shell
wget -O ~/.paloma/config/genesis.json https://raw.githubusercontent.com/palomachain/testnet/master/paloma-testnet-15/genesis.json
wget -O ~/.paloma/config/addrbook.json https://raw.githubusercontent.com/palomachain/testnet/master/paloma-testnet-15/addrbook.json
```

Mainnet:
```shell
wget -O ~/.paloma/config/genesis.json https://raw.githubusercontent.com/palomachain/mainnet/master/messenger/genesis.json
wget -O ~/.paloma/config/addrbook.json https://raw.githubusercontent.com/palomachain/mainnet/master/messenger/addrbook.json
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
palomad query bank balances --node tcp://testnet.palomaswap.com:26656 "$ADDRESS"
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
      --chain-id=$CHAIN_ID \
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
Environment=PIGEON_HEALTHCHECK_PORT=5757
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

