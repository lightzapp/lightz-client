# 🏅 Lightz Pro

Lightz Pro is a service designed to dynamically adjust swap fees based on Lightz's
liquidity needs, helping to maintain wallet and Lightning channel balances.

## Basics

Lightz Client is the recommended way to programmatically interact with Lightz Pro
to check for fee discounts, identify earn opportunities, and trigger swaps.

To configure Lightz Client to use the Lightz Pro API, simply start the daemon with
the `--pro` startup flag or set the `pro`
[configuration option](configuration.md). Since Lightz Pro discounts and earn
opportunities are primarily available for Chain -> Lightning swaps, this guide
will focus on that setup.

Lightz Client exposes a powerful [gRPC API](grpc.md), which you can integrate
into your own applications. For scripted usage of `lightzcli`, use the `--json`
flag, which is available on most commands.

The current fee rates can be retrieved using the [`GetPairs`](grpc.md#getpairs)
endpoint or with the `lightzcli getpairs` command.

Here is an example for querying the current service fee for a Bitcoin ->
Lightning swap using `lightzcli` and `jq` for processing the JSON output:

```bash
lightzcli getpairs --json | jq '.submarine[] | select(.pair.from == "BTC") | .fees.percentage'
```

## Paying Lightning Invoices

### **Paying invoices of your own node**

- [Connect Lightz Client to your CLN or LND node](index.md#configuration)
- Set the `amount` field to automatically generate a new invoice
- Example: `lightzcli createswap 100000 btc` (100k sats)

### **Paying invoices of an external service**

- [Start Lightz Client in standalone mode](index.md#standalone)
- Provide an existing invoice via `--invoice` or the `invoice` field
- Example: `lightzcli createswap --invoice lnbc1... btc`

## Funding Swaps

You can fund swaps in two ways:

1. Using Lightz Client's internal wallets
2. Using an external wallet

This choice is controlled by:

- API: `send_from_internal` parameter in
  [`CreateSwapRequest`](grpc.md#createswaprequest)
- CLI: `--from-wallet <wallet-name>` (internal) or `--external-pay` (external).
  If neither is specified, the first internal wallet with the correct currency
  will be used for funding.
