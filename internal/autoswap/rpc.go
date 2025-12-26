package autoswap

import (
	"github.com/lightzapp/lightz-client/internal/lightning"
	"github.com/lightzapp/lightz-client/pkg/lightzrpc"
	"github.com/lightzapp/lightz-client/pkg/lightzrpc/autoswaprpc"
	"github.com/lightzapp/lightz-client/pkg/lightzrpc/serializers"
)

func serializeLightningChannel(channel *lightning.LightningChannel) *lightzrpc.LightningChannel {
	if channel == nil {
		return nil
	}
	return &lightzrpc.LightningChannel{
		Id:          lightning.SerializeChanId(channel.Id),
		Capacity:    channel.Capacity,
		OutboundSat: channel.OutboundSat,
		InboundSat:  channel.InboundSat,
		PeerId:      channel.PeerId,
	}
}

func serializeLightningSwap(swap *LightningSwap) *autoswaprpc.LightningSwap {
	if swap == nil {
		return nil
	}
	return &autoswaprpc.LightningSwap{
		Amount:           swap.Amount,
		Type:             serializers.SerializeSwapType(swap.Type),
		FeeEstimate:      swap.FeeEstimate,
		DismissedReasons: swap.DismissedReasons,
	}
}

func serializeAutoChainSwap(swap *ChainSwap) *autoswaprpc.ChainSwap {
	if swap == nil {
		return nil
	}
	return &autoswaprpc.ChainSwap{
		Amount:           swap.Amount,
		FeeEstimate:      swap.FeeEstimate,
		DismissedReasons: swap.DismissedReasons,
	}
}
