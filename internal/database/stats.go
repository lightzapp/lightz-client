package database

import (
	"fmt"
	"strings"

	"github.com/lightzapp/lightz-client/pkg/lightz"
	"github.com/lightzapp/lightz-client/pkg/lightzrpc"
)

const statsQuery = `
SELECT COALESCE(SUM(COALESCE(serviceFee, 0) + COALESCE(onchainFee, 0)), 0),
       COALESCE(SUM(CASE WHEN state == 1 THEN amount END), 0),
       COUNT(*),
       COUNT(CASE WHEN state == 1 THEN 1 END)
FROM allSwaps
`

func (database *Database) QueryStats(args SwapQuery, swapTypes []lightz.SwapType) (*lightzrpc.SwapStats, error) {
	var placeholders []string
	var values []any
	for _, swapType := range swapTypes {
		placeholders = append(placeholders, "?")
		values = append(values, swapType)
	}
	where, values := args.ToWhereClauseWithExisting([]string{fmt.Sprintf("type IN (%s)", strings.Join(placeholders, ", "))}, values)

	rows := database.QueryRow(statsQuery+where, values...)

	stats := lightzrpc.SwapStats{}

	err := rows.Scan(&stats.TotalFees, &stats.TotalAmount, &stats.Count, &stats.SuccessCount)
	if err != nil {
		return nil, err
	}
	if stats.SuccessCount != 0 {
		stats.AvgFees = stats.TotalFees / int64(stats.SuccessCount)
		stats.AvgAmount = stats.TotalAmount / stats.SuccessCount
	}
	return &stats, nil
}
