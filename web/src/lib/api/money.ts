/**
 * Money conventions on the wire (mirrors the Go backend, DESIGN_PLAN §10.1).
 *
 * The backend uses a fixed-point `money.Price = int64` with `Scale = 1_000_000` (6 decimal places).
 * Integer money fields are therefore scaled by 1e6 — note the cashflow field named `amountCents` and
 * the analytics `*_cents` fields are a misnomer: they are 1e6-scaled, NOT hundredths. Other endpoints
 * (assets, portfolio transactions) return money as human-readable decimal strings (e.g. "1234.560000").
 */
export const MONEY_SCALE = 1_000_000;

/** Convert a 1e6-scaled integer money value (e.g. `amountCents`, `cost_basis`, EOD `Open`) to a number. */
export function scaledToNumber(scaled: number): number {
	return scaled / MONEY_SCALE;
}

/** Parse a decimal-string money value (e.g. assets `current_worth`, portfolio tx `amount`) to a number. */
export function decimalStringToNumber(value: string): number {
	const parsed = Number.parseFloat(value);
	return Number.isFinite(parsed) ? parsed : 0;
}

/** Convert a number back to the 1e6-scaled integer representation (for request payloads / round-trips). */
export function numberToScaled(value: number): number {
	return Math.round(value * MONEY_SCALE);
}
