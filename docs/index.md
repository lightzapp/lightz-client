---
next:
  text: "💰 Wallets"
  link: "/wallets"
---

# 👋 Introduction

Lightz Client connects to [CLN](https://github.com/ElementsProject/lightning/) or
[LND](https://github.com/flokiorg/flnd/) nodes and allows for fully
unattended channel rebalancing or accepting Lightning payments without running a
node, using [Lightz API](https://docs.lightz.exchange/v/api). It is composed of
`lightzcli` and `lightzd`.

## Design principles

- CLI-first: fine-grained control and enhanced setup UX via `lightzcli`
- CLN-first: first-class citizen support for
  [CLN](https://github.com/ElementsProject/lightning) in addition to
  [LND](https://github.com/flokiorg/flnd)
- [Liquid](https://liquid.net/)-first: leverages fee-efficient Lightning ->
  Liquid swaps in addition to Bitcoin swaps
- Self-contained: create/import Liquid and mainchain wallets, swap to read-only
  wallets

## Main Components

### `lightzd`

`lightzd` is a daemon that should run alongside of your lightning node. It
connects to your lightning node and the Lightz API to create and execute swaps.

### `lightzcli`

`lightzcli` is the CLI tool to interact with the gRPC interface that `lightzd`
exposes.

## Installation

### Binaries

Lightz Client is available for linux on `amd64` and `arm64`. Download the latest
binaries from the
[release](https://github.com/lightzapp/lightzd-client/releases) page. If you
are on another platform, use the docker image below.

### Docker

Lightz Client is also available as
[docker image](https://hub.docker.com/r/lightzd/lightzd-client/tags). Assuming your
lnd macaroon and certificate are placed in `~/.lnd`, run:

```bash
docker create -v ~/.lightzd:/root/.lightzd -v ~/.lnd:/root/.lnd --name lightzd-client lightzd/lightzd-client:latest
docker start lightzd-client
docker exec lightzd-client lightzcli getinfo
```

### Building from source

To build, you need [Go](https://go.dev/) ≥ 1.21, the Rust toolchain (including
[cargo](https://doc.rust-lang.org/cargo/) and `rustc`), and a C compiler such as
`gcc` for the native dependencies used by the daemon.

Lightz Client depends on [GDK](https://github.com/Blockstream/gdk) by
blockstream, which can be either dynamically or statically linked. The
recommended way to build from source is linking dynamically as a static link
requires compiling gdk as well.

First, initialise the required git submodules (e.g. `LWK`) with
`git submodule update --init --recursive`, then build the daemon and CLI with
`make build`. The binaries will be placed in the current directory.

You can also run `make install` which will place the binaries into your `GOBIN`
(`$HOME/go/bin` by default) directory.

## Setup

Binaries for the latest release of Lightz Client can be found on the
[release page](https://github.com/lightzapp/lightzd-client/releases). If no
binaries are available for your platform, you can build them yourself with the
instructions above.

### Configuration

Configuration can be done via CLI params or a TOML configuration file (by
default located in `~/.lightzd/lightz.toml`). We suggest starting off with the
sample configuration file, which can be found [here](configuration.md).

`lightzd` requires a connection to a lightning node, which can be CLN or LND. If
you set configuration values for both, you can specify which to use with the
`node` param.

To view all CLI flags use `--help`.

#### LND

The LND node to which the daemon connects has to be version `v0.10.0-beta` or
higher. Also, LND needs to be compiled with these build flags (official binaries
from Lightning Labs releases include them):

- `invoicerpc` (hold invoices)
- `routerrpc` (multi path payments)
- `chainrpc` (block listener)
- `walletrpc` (fee estimations)

By default, lightzd will attempt to connect to lnd running at `localhost:10009`
(`lnd.host` and `lnd.port`) and looking for certificate and macaroon inside the
data directory `~/.lnd` (`lnd.datadir`).

You can manually set the location of the tls certificate (`lnd.certificate`) and
admin macaroon (`lnd.macaroon`) instead of speciyfing the data directory as
well.

#### CLN

The daemon connects to CLN through
[gRPC](https://docs.corelightning.org/docs/grpc). You need start CLN with the
`--grpc-port` CLI flag, or set it in your config:

- `--cln.port` same port as used for `--grpc-port`
- `--cln.host` host of the machine CLN is running on
- `--cln.datadir` data directory of cln (`~/.lightning` by default)

You can manually set the paths of `cln.rootcert`, `cln.privatekey` and
`cln.certchain` instead of speciyfing the data directory as well. You might have
to set the `cln.servername` option as well, if you are using a custom
certificate.

#### Standalone

The daemon can also operate without a lightning node. In this case, you need to
specify the `--standalone` CLI flag or set the `standalone` option to `true` in
the configuration file.

### Swap Mnemonic

The swap mnemonic is used to derive the private keys for each swap. It is
recommended to back this mnemonic up in a secure location, as it can be used to
recover funds locked in swaps in the case the database of the daemon is lost.

On initial startup, the client will generate a new swap mnemonic and store it in
the database. It can be viewed with the `lightzcli swapmnemonic get` command or
changed using the `lightzcli swapmnemonic set` command.

### CLI

We recommend running `lightzcli completions` to setup autocompletions for the CLI
(only supported for zsh and bash).

### Macaroons

The macaroons for the gRPC server of `lightzd` can be found in the `macaroons`
folder inside the data directory of the daemon. By default, that data directory
is `~/.lightzd` on Linux.
