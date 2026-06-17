/**
 * Runtime data-source configuration (DESIGN_PLAN §10.2). Read from Vite env so it works identically
 * in the SvelteKit app and in Storybook's Vite. Mocks are ON by default so the portal runs entirely
 * standalone; provide `VITE_API_URL` (and optionally `VITE_USE_MOCKS=false`) to hit the real backend.
 */
const env = import.meta.env as Record<string, string | undefined>;

/** Base URL of the Go API (no trailing slash). Empty when running against mocks. */
export const apiBaseUrl: string = (env.VITE_API_URL ?? '').replace(/\/+$/, '');

/** Whether service calls resolve from fixtures (`true`) or hit the live API (`false`). */
export const useMocks: boolean = (() => {
	if (env.VITE_USE_MOCKS === 'true') return true;
	if (env.VITE_USE_MOCKS === 'false') return false;
	return apiBaseUrl === '';
})();

/**
 * Single demo account the fixtures are scoped to. The portfolio/assets endpoints require an
 * `account_id`; in mock mode every service uses this stable id.
 */
export const DEMO_ACCOUNT_ID = '11111111-1111-4111-8111-111111111111';

/** Simulated latency (ms) for mock responses so loading states are observable in dev/Storybook. */
export const MOCK_LATENCY_MS = 350;
