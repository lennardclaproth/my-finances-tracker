import { MOCK_LATENCY_MS } from '$lib/api/config';

/** Resolve after a short delay so loading states are observable when running on mocks. */
export function delay(ms: number = MOCK_LATENCY_MS): Promise<void> {
	return new Promise((resolve) => setTimeout(resolve, ms));
}

/** Defensive copy so callers can't mutate the shared fixture arrays. */
export function clone<T>(value: T): T {
	return typeof structuredClone === 'function'
		? structuredClone(value)
		: (JSON.parse(JSON.stringify(value)) as T);
}

/** Case-insensitive "contains" used by the in-memory text filters. */
export function contains(haystack: string | null | undefined, needle: string | undefined): boolean {
	if (!needle) return true;
	return (haystack ?? '').toLowerCase().includes(needle.toLowerCase());
}

/** A synthetic UUID for mock write responses (nothing is persisted in mock mode). */
export function mockId(): string {
	return typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
		? crypto.randomUUID()
		: '00000000-0000-4000-8000-000000000000';
}
