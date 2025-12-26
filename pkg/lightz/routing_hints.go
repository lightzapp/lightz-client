package lightz

import (
	"github.com/flokiorg/flnd/zpay32"
	btcec "github.com/flokiorg/go-flokicoin/crypto"
)

const magicRoutingHintConstant = 596385002596073472

func FindMagicRoutingHint(invoice *zpay32.Invoice) *btcec.PublicKey {
	for _, h := range invoice.RouteHints {
		for _, hint := range h {
			if hint.ChannelID == magicRoutingHintConstant {
				return hint.NodeID
			}
		}
	}
	return nil
}
