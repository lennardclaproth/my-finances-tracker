import { numberToScaled as s } from '$lib/api/money';
import type {
	CashflowTransaction,
	CashflowMonthlyPoint,
	TagDistributionEntry
} from '$lib/api/types';

/**
 * Deterministic cashflow transactions spanning early 2026. Mix of directions, tags, sources, a few
 * untagged and a couple ignored rows so table filtering / empty / pagination states are exercisable.
 * `amountCents` is the 1e6-scaled magnitude; sign is carried by `direction`.
 */
export const cashflowTransactions: CashflowTransaction[] = [
	{
		id: 'cf-0001',
		description: 'Monthly salary',
		note: 'ACME Corp payroll',
		source: 'ing',
		amountCents: s(4200),
		direction: 'in',
		date: '2026-06-25T08:00:00Z',
		tag: 'salary',
		ignored: false
	},
	{
		id: 'cf-0002',
		description: 'Albert Heijn',
		note: 'Weekly groceries',
		source: 'ing',
		amountCents: s(74.32),
		direction: 'out',
		date: '2026-06-22T18:14:00Z',
		tag: 'groceries',
		ignored: false
	},
	{
		id: 'cf-0003',
		description: 'Rent — June',
		note: 'Apartment',
		source: 'ing',
		amountCents: s(1450),
		direction: 'out',
		date: '2026-06-01T06:00:00Z',
		tag: 'rent',
		ignored: false
	},
	{
		id: 'cf-0004',
		description: 'Freelance invoice #42',
		note: 'Design work',
		source: 'n26',
		amountCents: s(1200),
		direction: 'in',
		date: '2026-06-18T11:30:00Z',
		tag: 'freelance',
		ignored: false
	},
	{
		id: 'cf-0005',
		description: 'Spotify',
		note: 'Subscription',
		source: 'n26',
		amountCents: s(11.99),
		direction: 'out',
		date: '2026-06-15T03:00:00Z',
		tag: 'subscriptions',
		ignored: false
	},
	{
		id: 'cf-0006',
		description: 'NS train',
		note: 'Commute',
		source: 'ing',
		amountCents: s(28.4),
		direction: 'out',
		date: '2026-06-14T07:45:00Z',
		tag: 'transport',
		ignored: false
	},
	{
		id: 'cf-0007',
		description: 'Dinner — Toscanini',
		note: '',
		source: 'n26',
		amountCents: s(96.5),
		direction: 'out',
		date: '2026-06-12T20:10:00Z',
		tag: 'dining',
		ignored: false
	},
	{
		id: 'cf-0008',
		description: 'Energy bill',
		note: 'Eneco',
		source: 'ing',
		amountCents: s(132.18),
		direction: 'out',
		date: '2026-06-10T09:00:00Z',
		tag: 'utilities',
		ignored: false
	},
	{
		id: 'cf-0009',
		description: 'Pharmacy',
		note: '',
		source: 'ing',
		amountCents: s(23.75),
		direction: 'out',
		date: '2026-06-09T16:20:00Z',
		tag: 'health',
		ignored: false
	},
	{
		id: 'cf-0010',
		description: 'Bol.com',
		note: 'Mystery charge',
		source: 'n26',
		amountCents: s(54.99),
		direction: 'out',
		date: '2026-06-08T13:05:00Z',
		tag: '',
		ignored: false
	},
	{
		id: 'cf-0011',
		description: 'Transfer to savings',
		note: '',
		source: 'ing',
		amountCents: s(500),
		direction: 'out',
		date: '2026-06-26T08:30:00Z',
		tag: 'savings',
		ignored: true
	},
	{
		id: 'cf-0012',
		description: 'Cinema',
		note: 'Pathé',
		source: 'n26',
		amountCents: s(31),
		direction: 'out',
		date: '2026-06-07T21:00:00Z',
		tag: 'entertainment',
		ignored: false
	},
	{
		id: 'cf-0013',
		description: 'Monthly salary',
		note: 'ACME Corp payroll',
		source: 'ing',
		amountCents: s(4200),
		direction: 'in',
		date: '2026-05-25T08:00:00Z',
		tag: 'salary',
		ignored: false
	},
	{
		id: 'cf-0014',
		description: 'Jumbo',
		note: 'Groceries',
		source: 'ing',
		amountCents: s(88.6),
		direction: 'out',
		date: '2026-05-21T17:40:00Z',
		tag: 'groceries',
		ignored: false
	},
	{
		id: 'cf-0015',
		description: 'Rent — May',
		note: 'Apartment',
		source: 'ing',
		amountCents: s(1450),
		direction: 'out',
		date: '2026-05-01T06:00:00Z',
		tag: 'rent',
		ignored: false
	},
	{
		id: 'cf-0016',
		description: 'Freelance invoice #41',
		note: 'Consulting',
		source: 'n26',
		amountCents: s(900),
		direction: 'in',
		date: '2026-05-19T10:00:00Z',
		tag: 'freelance',
		ignored: false
	},
	{
		id: 'cf-0017',
		description: 'Water bill',
		note: 'Vitens',
		source: 'ing',
		amountCents: s(41.2),
		direction: 'out',
		date: '2026-05-11T09:00:00Z',
		tag: 'utilities',
		ignored: false
	},
	{
		id: 'cf-0018',
		description: 'Coffee — Anne&Max',
		note: '',
		source: 'n26',
		amountCents: s(8.5),
		direction: 'out',
		date: '2026-05-09T08:20:00Z',
		tag: 'dining',
		ignored: false
	},
	{
		id: 'cf-0019',
		description: 'Refund — webshop',
		note: 'Returned item',
		source: 'n26',
		amountCents: s(45),
		direction: 'in',
		date: '2026-05-06T12:00:00Z',
		tag: '',
		ignored: false
	},
	{
		id: 'cf-0020',
		description: 'Gym membership',
		note: 'Basic-Fit',
		source: 'ing',
		amountCents: s(24.99),
		direction: 'out',
		date: '2026-05-03T05:00:00Z',
		tag: 'health',
		ignored: false
	},
	{
		id: 'cf-0021',
		description: 'Monthly salary',
		note: 'ACME Corp payroll',
		source: 'ing',
		amountCents: s(4200),
		direction: 'in',
		date: '2026-04-25T08:00:00Z',
		tag: 'salary',
		ignored: false
	},
	{
		id: 'cf-0022',
		description: 'Rent — April',
		note: 'Apartment',
		source: 'ing',
		amountCents: s(1450),
		direction: 'out',
		date: '2026-04-01T06:00:00Z',
		tag: 'rent',
		ignored: false
	},
	{
		id: 'cf-0023',
		description: 'Groceries — Lidl',
		note: '',
		source: 'ing',
		amountCents: s(63.18),
		direction: 'out',
		date: '2026-04-15T18:00:00Z',
		tag: 'groceries',
		ignored: false
	},
	{
		id: 'cf-0024',
		description: 'Electricity',
		note: 'Eneco',
		source: 'ing',
		amountCents: s(118.04),
		direction: 'out',
		date: '2026-04-10T09:00:00Z',
		tag: 'utilities',
		ignored: false
	},
	{
		id: 'cf-0025',
		description: 'Concert tickets',
		note: 'Ziggo Dome',
		source: 'n26',
		amountCents: s(140),
		direction: 'out',
		date: '2026-04-08T19:30:00Z',
		tag: 'entertainment',
		ignored: false
	},
	{
		id: 'cf-0026',
		description: 'Dividend payout',
		note: 'Broker',
		source: 'degiro',
		amountCents: s(62.4),
		direction: 'in',
		date: '2026-04-05T10:00:00Z',
		tag: 'investments',
		ignored: false
	},
	{
		id: 'cf-0027',
		description: 'Taxi',
		note: '',
		source: 'n26',
		amountCents: s(19.8),
		direction: 'out',
		date: '2026-04-03T23:10:00Z',
		tag: 'transport',
		ignored: false
	},
	{
		id: 'cf-0028',
		description: 'Bookstore',
		note: 'Gift',
		source: 'n26',
		amountCents: s(34.5),
		direction: 'out',
		date: '2026-03-28T15:00:00Z',
		tag: '',
		ignored: false
	},
	{
		id: 'cf-0029',
		description: 'Monthly salary',
		note: 'ACME Corp payroll',
		source: 'ing',
		amountCents: s(4200),
		direction: 'in',
		date: '2026-03-25T08:00:00Z',
		tag: 'salary',
		ignored: false
	},
	{
		id: 'cf-0030',
		description: 'Rent — March',
		note: 'Apartment',
		source: 'ing',
		amountCents: s(1450),
		direction: 'out',
		date: '2026-03-01T06:00:00Z',
		tag: 'rent',
		ignored: false
	}
];

/** Monthly incoming/outgoing/net totals (1e6-scaled). First-of-month keys, oldest → newest. */
export const cashflowMonthly: CashflowMonthlyPoint[] = [
	{ month: '2026-01-01', incoming_cents: s(5100), outgoing_cents: s(3420), net_cents: s(1680) },
	{ month: '2026-02-01', incoming_cents: s(4200), outgoing_cents: s(3890), net_cents: s(310) },
	{ month: '2026-03-01', incoming_cents: s(4200), outgoing_cents: s(3650), net_cents: s(550) },
	{
		month: '2026-04-01',
		incoming_cents: s(4262.4),
		outgoing_cents: s(4001.05),
		net_cents: s(261.35)
	},
	{
		month: '2026-05-01',
		incoming_cents: s(5145),
		outgoing_cents: s(3127.28),
		net_cents: s(2017.72)
	},
	{ month: '2026-06-01', incoming_cents: s(5400), outgoing_cents: s(3815.5), net_cents: s(1584.5) }
];

/** Tag totals (1e6-scaled) for the incoming/outgoing donuts and the combined view. */
export const cashflowTagDistribution: {
	combined: TagDistributionEntry[];
	incoming: TagDistributionEntry[];
	outgoing: TagDistributionEntry[];
} = {
	incoming: [
		{ tag: 'salary', totalCents: s(12600) },
		{ tag: 'freelance', totalCents: s(2100) },
		{ tag: 'investments', totalCents: s(62.4) },
		{ tag: '', totalCents: s(45) }
	],
	outgoing: [
		{ tag: 'rent', totalCents: s(5800) },
		{ tag: 'groceries', totalCents: s(389.5) },
		{ tag: 'utilities', totalCents: s(291.42) },
		{ tag: 'dining', totalCents: s(105) },
		{ tag: 'entertainment', totalCents: s(311) },
		{ tag: 'transport', totalCents: s(48.2) },
		{ tag: 'health', totalCents: s(48.74) },
		{ tag: 'subscriptions', totalCents: s(11.99) },
		{ tag: '', totalCents: s(89.49) }
	],
	combined: [
		{ tag: 'salary', totalCents: s(12600) },
		{ tag: 'rent', totalCents: s(5800) },
		{ tag: 'freelance', totalCents: s(2100) },
		{ tag: 'groceries', totalCents: s(389.5) },
		{ tag: 'entertainment', totalCents: s(311) },
		{ tag: 'utilities', totalCents: s(291.42) },
		{ tag: 'dining', totalCents: s(105) }
	]
};
