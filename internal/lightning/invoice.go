package lightning

import "C"
import (
	"time"

	"github.com/flokiorg/flnd/zpay32"
	"github.com/flokiorg/go-flokicoin/chaincfg"
	btcec "github.com/flokiorg/go-flokicoin/crypto"
	"github.com/lightzapp/lightz-client/pkg/lightz"
)

type DecodedInvoice struct {
	AmountSat        uint64
	PaymentHash      [32]byte
	Expiry           time.Time
	MagicRoutingHint *btcec.PublicKey
}

func DecodeInvoice(invoice string, network *chaincfg.Params) (*DecodedInvoice, error) {
	bolt11, err := zpay32.Decode(invoice, network)
	if err == nil {
		var amount uint64
		if bolt11.MilliSat != nil {
			amount = uint64(bolt11.MilliSat.ToLokis())
		}
		return &DecodedInvoice{
			AmountSat:        amount,
			PaymentHash:      *bolt11.PaymentHash,
			Expiry:           bolt11.Timestamp.Add(bolt11.Expiry()),
			MagicRoutingHint: lightz.FindMagicRoutingHint(bolt11),
		}, nil
	}
	return DecodeBolt12Invoice(invoice)
}
