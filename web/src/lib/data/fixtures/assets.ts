import type { AssetClass, AssetClassDetails, AssetSnapshotPoint } from '$lib/api/types';

/** Asset classes for the account. Money is a decimal string; `growth_pct` is a percentage number. */
export const assetClasses: AssetClass[] = [
	{
		id: 'cls-cash',
		name: 'Cash & savings',
		source: 'manual',
		archived: false,
		current_worth: '18450.000000',
		last_change_at: '2026-06-12T09:00:00Z',
		growth_pct: 2.1,
		updated_at: '2026-06-12T09:00:00Z'
	},
	{
		id: 'cls-stocks',
		name: 'Brokerage',
		source: 'degiro',
		archived: false,
		current_worth: '12230.400000',
		last_change_at: '2026-06-16T22:00:00Z',
		growth_pct: 12.31,
		updated_at: '2026-06-16T22:00:00Z'
	},
	{
		id: 'cls-pension',
		name: 'Pension (BND)',
		source: 'brandnewday',
		archived: false,
		current_worth: '34120.500000',
		last_change_at: '2026-06-01T06:00:00Z',
		growth_pct: 5.4,
		updated_at: '2026-06-01T06:00:00Z'
	},
	{
		id: 'cls-crypto',
		name: 'Crypto',
		source: 'manual',
		archived: false,
		current_worth: '2980.250000',
		last_change_at: '2026-06-15T20:00:00Z',
		growth_pct: -8.7,
		updated_at: '2026-06-15T20:00:00Z'
	},
	{
		id: 'cls-realestate',
		name: 'Real estate equity',
		source: 'manual',
		archived: false,
		current_worth: '85000.000000',
		last_change_at: '2026-04-01T06:00:00Z',
		growth_pct: null,
		updated_at: '2026-04-01T06:00:00Z'
	},
	{
		id: 'cls-old',
		name: 'Legacy savings (archived)',
		source: 'manual',
		archived: true,
		current_worth: '0.000000',
		last_change_at: '2025-11-30T09:00:00Z',
		growth_pct: null,
		updated_at: '2025-11-30T09:00:00Z'
	}
];

/** Six-point growth timeline for a class (decimal-string total worth, oldest → newest). */
function growth(values: number[]): { date: string; total_worth: string }[] {
	const months = [
		'2026-01-01',
		'2026-02-01',
		'2026-03-01',
		'2026-04-01',
		'2026-05-01',
		'2026-06-01'
	];
	return values.map((v, i) => ({ date: months[i], total_worth: v.toFixed(6) }));
}

/** Class detail (drawer) fixtures keyed by class id. */
export const assetClassDetails: Record<string, AssetClassDetails> = {
	'cls-cash': {
		class: assetClasses[0],
		assets: [
			{
				id: 'ast-ing-savings',
				name: 'ING savings',
				current_worth: '12450.000000',
				archived: false,
				updated_at: '2026-06-12T09:00:00Z'
			},
			{
				id: 'ast-n26-space',
				name: 'N26 space',
				current_worth: '6000.000000',
				archived: false,
				updated_at: '2026-06-10T09:00:00Z'
			}
		],
		growth: growth([16800, 17100, 17500, 17900, 18200, 18450]),
		mutations: [
			{
				id: 'mut-c1',
				item_id: 'ast-ing-savings',
				change_type: 'deposit',
				direction: 'in',
				amount: '500.000000',
				previous_worth: '11950.000000',
				new_worth: '12450.000000',
				class_total_worth: '18450.000000',
				effective_date: '2026-06-12',
				note: 'Monthly transfer',
				created_at: '2026-06-12T09:00:00Z'
			},
			{
				id: 'mut-c2',
				item_id: 'ast-n26-space',
				change_type: 'deposit',
				direction: 'in',
				amount: '250.000000',
				previous_worth: '5750.000000',
				new_worth: '6000.000000',
				class_total_worth: '17950.000000',
				effective_date: '2026-05-12',
				note: null,
				created_at: '2026-05-12T09:00:00Z'
			}
		]
	},
	'cls-stocks': {
		class: assetClasses[1],
		assets: [
			{
				id: 'ast-vwrl',
				name: 'VWRL',
				current_worth: '4973.400000',
				archived: false,
				updated_at: '2026-06-16T22:00:00Z'
			},
			{
				id: 'ast-iwda',
				name: 'IWDA',
				current_worth: '3105.000000',
				archived: false,
				updated_at: '2026-06-16T22:00:00Z'
			},
			{
				id: 'ast-aapl',
				name: 'AAPL',
				current_worth: '2256.000000',
				archived: false,
				updated_at: '2026-06-16T22:00:00Z'
			},
			{
				id: 'ast-asml',
				name: 'ASML',
				current_worth: '1896.000000',
				archived: false,
				updated_at: '2026-06-16T22:00:00Z'
			}
		],
		growth: growth([9000, 9610, 9740, 10120, 10790, 12230.4]),
		mutations: [
			{
				id: 'mut-s1',
				item_id: 'ast-vwrl',
				change_type: 'valuation',
				direction: null,
				amount: '180.400000',
				previous_worth: '4793.000000',
				new_worth: '4973.400000',
				class_total_worth: '12230.400000',
				effective_date: '2026-06-16',
				note: 'EOD revaluation',
				created_at: '2026-06-16T22:00:00Z'
			}
		]
	}
};

/** Account-level total worth timeline (decimal-string), oldest → newest. */
export const assetSnapshots: AssetSnapshotPoint[] = [
	{ date: '2026-01-01', total_worth: '142300.000000' },
	{ date: '2026-02-01', total_worth: '144980.000000' },
	{ date: '2026-03-01', total_worth: '146120.000000' },
	{ date: '2026-04-01', total_worth: '149870.000000' },
	{ date: '2026-05-01', total_worth: '151540.000000' },
	{ date: '2026-06-01', total_worth: '152781.150000' }
];
