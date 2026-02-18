const euroFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "EUR",
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});
const MONEY_SCALE = 1_000_000;

const localDateFormatter = new Intl.DateTimeFormat(undefined, {
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
});

export function formatAmountCents(amountCents: number, direction: string): string {
  const amount = Math.abs(amountCents) / MONEY_SCALE;
  const normalizedDirection = direction.trim().toLowerCase();
  const signedAmount = normalizedDirection === "out" ? -amount : amount;
  return euroFormatter.format(signedAmount);
}

export function formatLocalDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return localDateFormatter.format(date);
}

export function normalizeDirection(value: string): string {
  const normalized = value.trim().toLowerCase();
  if (normalized === "in") {
    return "In";
  }
  if (normalized === "out") {
    return "Out";
  }
  return value;
}
