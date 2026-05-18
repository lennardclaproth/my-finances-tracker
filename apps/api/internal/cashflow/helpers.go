package cashflow

import "time"

func manualCashflowRowNumber(idx int) int {
	const maxRowNumber = 2_147_483_647
	row := int(time.Now().UnixNano()%2_000_000_000) + idx + 1
	if row <= 0 {
		return idx + 1
	}
	if row > maxRowNumber {
		return (row % maxRowNumber) + 1
	}
	return row
}
