import { apiBaseUrl } from './config';

/** A query value; arrays are joined with commas (matches the backend's comma-separated params). */
type QueryValue = string | number | boolean | string[] | undefined | null;

/** Error thrown when the API responds with a non-2xx status. */
export class ApiError extends Error {
	readonly status: number;
	readonly body: unknown;

	constructor(status: number, message: string, body: unknown) {
		super(message);
		this.name = 'ApiError';
		this.status = status;
		this.body = body;
	}
}

function buildQuery(query?: Record<string, QueryValue>): string {
	if (!query) return '';
	const params = new URLSearchParams();
	for (const [key, value] of Object.entries(query)) {
		if (value === undefined || value === null || value === '') continue;
		params.set(key, Array.isArray(value) ? value.join(',') : String(value));
	}
	const qs = params.toString();
	return qs ? `?${qs}` : '';
}

async function parse<T>(res: Response): Promise<T> {
	const text = await res.text();
	const body = text ? (JSON.parse(text) as unknown) : null;
	if (!res.ok) {
		throw new ApiError(res.status, `Request failed with status ${res.status}`, body);
	}
	return body as T;
}

/** GET `path` with optional query params, returning the decoded JSON body. */
export async function apiGet<T>(path: string, query?: Record<string, QueryValue>): Promise<T> {
	const res = await fetch(`${apiBaseUrl}${path}${buildQuery(query)}`, {
		headers: { Accept: 'application/json' }
	});
	return parse<T>(res);
}

/** Send a JSON body with the given method (POST/PUT/PATCH/DELETE), returning the decoded JSON body. */
export async function apiSend<T>(method: string, path: string, body?: unknown): Promise<T> {
	const res = await fetch(`${apiBaseUrl}${path}`, {
		method,
		headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
		body: body === undefined ? undefined : JSON.stringify(body)
	});
	return parse<T>(res);
}
