package autoswap

import (
	"errors"
	"time"

	"github.com/lightzapp/lightz-client/pkg/lightz"
	"github.com/lightzapp/lightz-client/pkg/lightzrpc"

	"github.com/lightzapp/lightz-client/internal/database"
)

type Budget struct {
	database.BudgetInterval
	Amount uint64
	Total  uint64
	Stats  *lightzrpc.SwapStats
}

type budgetConfig interface {
	GetBudgetInterval() uint64
	GetBudget() uint64
}

func (c *shared) GetCurrentBudget(
	createIfMissing bool,
	swapperType SwapperType,
	cfg budgetConfig,
	tenantId database.Id,
) (*Budget, error) {
	budgetDuration := time.Duration(cfg.GetBudgetInterval()) * time.Second
	totalBudget := cfg.GetBudget()

	currentInterval, err := c.database.QueryCurrentBudgetInterval(string(swapperType), tenantId)
	if err != nil {
		return nil, errors.New("Could not get budget period: " + err.Error())
	}

	now := time.Now()
	if currentInterval == nil || now.After(currentInterval.EndDate) {
		if createIfMissing {
			if currentInterval == nil {
				currentInterval = &database.BudgetInterval{
					StartDate: now,
					EndDate:   now.Add(budgetDuration),
					Name:      string(swapperType),
					TenantId:  tenantId,
				}
			}
			for now.After(currentInterval.EndDate) {
				currentInterval.StartDate = currentInterval.EndDate
				currentInterval.EndDate = currentInterval.EndDate.Add(budgetDuration)
			}
			if err := c.database.CreateBudget(*currentInterval); err != nil {
				return nil, errors.New("Could not create budget period: " + err.Error())
			}
		} else {
			return nil, nil
		}
	}

	var swapTypes []lightz.SwapType
	if swapperType == Lightning {
		swapTypes = []lightz.SwapType{lightz.NormalSwap, lightz.ReverseSwap}
	} else {
		swapTypes = []lightz.SwapType{lightz.ChainSwap}
	}
	stats, err := c.database.QueryStats(database.SwapQuery{
		Since:    currentInterval.StartDate,
		Include:  lightzrpc.IncludeSwaps_AUTO,
		TenantId: &tenantId,
	}, swapTypes)
	if err != nil {
		return nil, errors.New("Could not get past fees: " + err.Error())
	}

	budget := max(0, int64(totalBudget)-int64(stats.TotalFees))

	return &Budget{
		BudgetInterval: *currentInterval,
		Amount:         uint64(budget),
		Total:          totalBudget,
		Stats:          stats,
	}, nil
}
