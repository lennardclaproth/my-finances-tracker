/** Values captured by the cashflow transaction form modal. */
export interface CashflowTransactionFormValue {
	/** "YYYY-MM-DD". */
	date: string;
	/** Decimal-string amount. */
	amount: string;
	type: 'income' | 'expense';
	description: string;
	note: string;
	tag: string;
}
