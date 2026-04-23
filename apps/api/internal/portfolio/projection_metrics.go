package portfolio

// SnapshotReturnVsCostBasisPct computes total PnL percentage over cost basis.
func SnapshotReturnVsCostBasisPct(snapshot *PortfolioSnapshot) float64 {
	if snapshot == nil || snapshot.CostBasis == 0 {
		return 0
	}
	return (snapshot.TotalPnL.Float64() / snapshot.CostBasis.Float64()) * 100
}

// SnapshotValueIndex computes the index representation for a snapshot return.
func SnapshotValueIndex(snapshot *PortfolioSnapshot) float64 {
	if snapshot == nil {
		return 100
	}
	return 100.0 * (1 + (snapshot.TimeWeightedReturnPct / 100))
}

// SignedAmountForRead normalizes transaction amount sign for read-model responses.
func SignedAmountForRead(txType TransactionType, quantity, amount float64) float64 {
	if txType == TxCash && quantity < 0 {
		return -amount
	}
	return amount
}
