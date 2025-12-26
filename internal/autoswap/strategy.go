package autoswap

import (
	"math"

	"github.com/lightzapp/lightz-client/internal/lightning"
	"github.com/lightzapp/lightz-client/internal/logger"
	"github.com/lightzapp/lightz-client/pkg/lightz"
	"github.com/lightzapp/lightz-client/pkg/lightzrpc/autoswaprpc"
)

type Strategy = func(channels []*lightning.LightningChannel) []*LightningRecommendation

func (cfg *LightningConfig) channelRecommendation(channel *lightning.LightningChannel) *LightningRecommendation {
	outbound := cfg.outboundBalance.Get(channel.Capacity)
	inbound := cfg.inboundBalance.Get(channel.Capacity)

	if channel.Capacity < outbound+inbound {
		logger.Warnf("Capacity of channel %d is smaller than the sum of the outbound and inbound tresholds", channel.Id)
		return nil
	}

	thresholds := &autoswaprpc.LightningThresholds{}
	if outbound > 0 {
		thresholds.Outbound = &outbound
	}
	if inbound > 0 {
		thresholds.Inbound = &inbound
	}
	recommendation := &LightningRecommendation{Channel: channel, Thresholds: thresholds}
	swap := &LightningSwap{}
	if channel.OutboundSat < outbound {
		swap.Type = lightz.NormalSwap
		if cfg.swapType == lightz.NormalSwap {
			swap.Amount = channel.InboundSat
		}
	} else if channel.InboundSat < inbound {
		swap.Type = lightz.ReverseSwap
		if cfg.swapType == lightz.ReverseSwap {
			swap.Amount = channel.OutboundSat
		}
	}
	if swap.Type != "" && cfg.Allowed(swap.Type) {
		if swap.Amount == 0 {
			target := float64(outbound+(channel.Capacity-inbound)) / 2
			swap.Amount = uint64(math.Abs(float64(channel.OutboundSat) - target))
		} else {
			reserve := lightz.CalculatePercentage(cfg.reserve, channel.Capacity)
			if swap.Amount < reserve {
				logger.Warnf(
					"Recommended amount %d of channel %d is lower than the reserve %d, not recommending swap",
					swap.Amount, channel.Id, reserve,
				)
				swap = nil
			} else {
				swap.Amount -= reserve
			}
		}
		recommendation.Swap = swap
	}
	return recommendation
}

func (cfg *LightningConfig) totalBalanceStrategy(channels []*lightning.LightningChannel) []*LightningRecommendation {
	var total lightning.LightningChannel

	for _, channel := range channels {
		total.OutboundSat += channel.OutboundSat
		total.InboundSat += channel.InboundSat
		total.Capacity += channel.Capacity
	}

	logger.Debugf("Total channel balances %+v", total)

	return []*LightningRecommendation{cfg.channelRecommendation(&total)}
}

func (cfg *LightningConfig) perChannelStrategy(channels []*lightning.LightningChannel) []*LightningRecommendation {
	var recommendations []*LightningRecommendation

	for _, channel := range channels {
		recommendations = append(recommendations, cfg.channelRecommendation(channel))
	}

	return recommendations
}
