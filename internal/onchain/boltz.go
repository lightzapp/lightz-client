package onchain

import (
	"errors"

	"github.com/lightzapp/lightz-client/pkg/lightz"
)

type LightzProvider struct {
	*lightz.Api
	currency lightz.Currency
}

var _ ChainProvider = &LightzProvider{}

func NewLightzChainProvider(lightzd *lightz.Api, currency lightz.Currency) *LightzProvider {
	return &LightzProvider{lightzd, currency}
}

func (txProvider LightzProvider) GetRawTransaction(txId string) (string, error) {
	return txProvider.GetTransaction(txId, txProvider.currency)
}

func (txProvider LightzProvider) BroadcastTransaction(txHex string) (string, error) {
	return txProvider.Api.BroadcastTransaction(txProvider.currency, txHex)
}

func (txProvider LightzProvider) IsTransactionConfirmed(txId string) (bool, error) {
	transaction, err := txProvider.GetTransactionDetails(txId, txProvider.currency)
	if err != nil {
		return false, err
	}
	return transaction.Confirmations > 0, nil
}

func (txProvider LightzProvider) EstimateFee() (float64, error) {
	return txProvider.Api.EstimateFee(txProvider.currency)
}

func (txProvider LightzProvider) GetBlockHeight() (uint32, error) {
	return txProvider.Api.GetBlockHeight(txProvider.currency)
}

func (txProvider LightzProvider) GetUnspentOutputs(address string) ([]*Output, error) {
	return nil, errors.ErrUnsupported
}

func (txProvider LightzProvider) Disconnect() {}
