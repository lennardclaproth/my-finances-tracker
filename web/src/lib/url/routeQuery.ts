/**
 * URL-as-state helpers (DESIGN_PLAN §5.8). Pure functions over URLSearchParams driven by a small
 * schema, so pages can parse `page.url.searchParams` into typed filter state and serialize it back —
 * omitting defaults so the URL stays clean. Ported from the reference's `routeQuery.ts`.
 */
export type QueryFieldType = 'string' | 'number' | 'boolean' | 'string[]';

export interface QueryField {
	type: QueryFieldType;
}

export type QuerySchema = Record<string, QueryField>;

export type QueryValue = string | number | boolean | string[];
export type QueryState = Record<string, QueryValue>;

/** Parse a URLSearchParams into typed state per the schema (missing values become type defaults). */
export function parseQuery(params: URLSearchParams, schema: QuerySchema): QueryState {
	const state: QueryState = {};
	for (const [key, field] of Object.entries(schema)) {
		const raw = params.get(key);
		switch (field.type) {
			case 'string':
				state[key] = raw ?? '';
				break;
			case 'number': {
				const parsed = raw === null ? 0 : Number(raw);
				state[key] = Number.isFinite(parsed) ? parsed : 0;
				break;
			}
			case 'boolean':
				state[key] = raw === 'true' || raw === '1';
				break;
			case 'string[]':
				state[key] = raw
					? raw
							.split(',')
							.map((part) => part.trim())
							.filter(Boolean)
					: [];
				break;
		}
	}
	return state;
}

/** Serialize typed state back to URLSearchParams, omitting empty/default values. */
export function serializeQuery(state: QueryState, schema: QuerySchema): URLSearchParams {
	const params = new URLSearchParams();
	for (const [key, field] of Object.entries(schema)) {
		const value = state[key];
		if (value === undefined || value === null) continue;
		switch (field.type) {
			case 'string':
				if (value !== '') params.set(key, String(value));
				break;
			case 'number':
				if (Number(value) !== 0) params.set(key, String(value));
				break;
			case 'boolean':
				if (value === true) params.set(key, 'true');
				break;
			case 'string[]':
				if (Array.isArray(value) && value.length > 0) params.set(key, value.join(','));
				break;
		}
	}
	return params;
}

function valuesEqual(a: QueryValue | undefined, b: QueryValue | undefined): boolean {
	if (Array.isArray(a) && Array.isArray(b)) {
		return a.length === b.length && a.every((value, index) => value === b[index]);
	}
	return a === b;
}

/** Stable string form of params (sorted) for equality checks. */
function canonical(params: URLSearchParams): string {
	const entries = [...params.entries()].sort((a, b) =>
		a[0] === b[0] ? (a[1] < b[1] ? -1 : 1) : a[0] < b[0] ? -1 : 1
	);
	return entries.map(([key, value]) => `${key}=${value}`).join('&');
}

/** Whether two query strings represent the same state (order-independent). */
export function areRouteQueriesEqual(a: URLSearchParams, b: URLSearchParams): boolean {
	return canonical(a) === canonical(b);
}

/** Whether `state` differs from `defaults` for any non-ignored key (drives the "filters active" badge). */
export function hasActiveFilters(
	state: QueryState,
	defaults: QueryState,
	ignore: string[] = []
): boolean {
	for (const key of Object.keys(state)) {
		if (ignore.includes(key)) continue;
		if (!valuesEqual(state[key], defaults[key])) return true;
	}
	return false;
}
