package serializers

import (
	"github.com/lightzapp/lightz-client/internal/onchain"
	"github.com/lightzapp/lightz-client/pkg/lightz"
	"github.com/lightzapp/lightz-client/pkg/lightzrpc"
)

func ParseCurrency(grpcCurrency *lightzrpc.Currency) lightz.Currency {
	if grpcCurrency == nil {
		return ""
	} else if *grpcCurrency == lightzrpc.Currency_BTC {
		return lightz.CurrencyBtc
	} else {
		return lightz.CurrencyLiquid
	}
}

func ParsePair(grpcPair *lightzrpc.Pair) (pair lightz.Pair) {
	if grpcPair == nil {
		return lightz.PairBtc
	}
	return lightz.Pair{
		From: ParseCurrency(&grpcPair.From),
		To:   ParseCurrency(&grpcPair.To),
	}
}

func SerializeCurrency(currency lightz.Currency) lightzrpc.Currency {
	if currency == lightz.CurrencyBtc {
		return lightzrpc.Currency_BTC
	} else {
		return lightzrpc.Currency_LBTC
	}
}

func SerializeSwapType(currency lightz.SwapType) lightzrpc.SwapType {
	switch currency {
	case lightz.NormalSwap:
		return lightzrpc.SwapType_SUBMARINE
	case lightz.ReverseSwap:
		return lightzrpc.SwapType_REVERSE
	default:
		return lightzrpc.SwapType_CHAIN
	}
}

func SerializePair(pair lightz.Pair) *lightzrpc.Pair {
	return &lightzrpc.Pair{
		From: SerializeCurrency(pair.From),
		To:   SerializeCurrency(pair.To),
	}
}

func SerializeWalletBalance(balance *onchain.Balance) *lightzrpc.Balance {
	if balance == nil {
		return nil
	}
	return &lightzrpc.Balance{
		Confirmed:   balance.Confirmed,
		Total:       balance.Total,
		Unconfirmed: balance.Unconfirmed,
	}
}
