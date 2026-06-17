/**
 * Pure calendar date math for the date pickers. Everything is UTC and uses "YYYY-MM-DD" strings to
 * match the backend's date params (`from`/`to`), so lexicographic string comparison equals date
 * comparison. No dependencies.
 */

/** One cell in a month grid. */
export interface CalendarDay {
	/** "YYYY-MM-DD" */
	iso: string;
	day: number;
	/** Belongs to the displayed month (vs. a leading/trailing spill day). */
	inMonth: boolean;
	isToday: boolean;
}

/** Monday-first weekday headers. */
export const weekdayLabels = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'] as const;

const monthNames = [
	'January',
	'February',
	'March',
	'April',
	'May',
	'June',
	'July',
	'August',
	'September',
	'October',
	'November',
	'December'
];

function pad(n: number): string {
	return n < 10 ? `0${n}` : String(n);
}

/** Format a Date as a UTC "YYYY-MM-DD" string. */
export function toISODate(date: Date): string {
	return `${date.getUTCFullYear()}-${pad(date.getUTCMonth() + 1)}-${pad(date.getUTCDate())}`;
}

/** Parse a "YYYY-MM-DD" string into a Date at UTC midnight. Returns null for invalid input. */
export function parseISODate(iso: string | null | undefined): Date | null {
	if (!iso) return null;
	const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(iso);
	if (!match) return null;
	const [, y, m, d] = match;
	return new Date(Date.UTC(Number(y), Number(m) - 1, Number(d)));
}

/** Today as a UTC "YYYY-MM-DD" string. */
export function todayISO(): string {
	return toISODate(new Date());
}

/** First day of the month containing `date`, at UTC midnight. */
export function startOfMonthUTC(date: Date): Date {
	return new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), 1));
}

/** Add `n` months (can be negative) to `date`, anchored to the first of the month. */
export function addMonthsUTC(date: Date, n: number): Date {
	return new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth() + n, 1));
}

/** Human label for a month, e.g. "June 2026". */
export function monthLabel(date: Date): string {
	return `${monthNames[date.getUTCMonth()]} ${date.getUTCFullYear()}`;
}

const shortMonthNames = monthNames.map((m) => m.slice(0, 3));

/** Human label for a single date, e.g. "12 Jun 2026". Empty string for invalid input. */
export function formatDisplayDate(iso: string | null | undefined): string {
	const date = parseISODate(iso);
	if (!date) return '';
	return `${date.getUTCDate()} ${shortMonthNames[date.getUTCMonth()]} ${date.getUTCFullYear()}`;
}

/**
 * Build a 42-cell (6×7) Monday-first grid for the month containing `month`, including leading/trailing
 * spill days from adjacent months.
 */
export function buildMonthGrid(month: Date): CalendarDay[] {
	const first = startOfMonthUTC(month);
	// JS getUTCDay: 0=Sun..6=Sat. Convert to Monday-first offset (Mon=0..Sun=6).
	const mondayOffset = (first.getUTCDay() + 6) % 7;
	const gridStart = new Date(first);
	gridStart.setUTCDate(first.getUTCDate() - mondayOffset);

	const today = todayISO();
	const cells: CalendarDay[] = [];
	for (let i = 0; i < 42; i++) {
		const date = new Date(gridStart);
		date.setUTCDate(gridStart.getUTCDate() + i);
		const iso = toISODate(date);
		cells.push({
			iso,
			day: date.getUTCDate(),
			inMonth: date.getUTCMonth() === month.getUTCMonth(),
			isToday: iso === today
		});
	}
	return cells;
}

/** Return `iso` shifted by `days` (can be negative) as a "YYYY-MM-DD" string. */
export function addDaysISO(iso: string, days: number): string {
	const base = parseISODate(iso) ?? new Date();
	return toISODate(new Date(base.getTime() + days * 86_400_000));
}

/** Whether `iso` falls within the inclusive [start, end] range (any bound may be null). */
export function isWithinRange(
	iso: string,
	start: string | null | undefined,
	end: string | null | undefined
): boolean {
	if (!start || !end) return false;
	const [lo, hi] = start <= end ? [start, end] : [end, start];
	return iso >= lo && iso <= hi;
}

/** Whether `iso` is selectable given optional min/max bounds. */
export function isDisabled(
	iso: string,
	min: string | null | undefined,
	max: string | null | undefined
): boolean {
	if (min && iso < min) return true;
	if (max && iso > max) return true;
	return false;
}
