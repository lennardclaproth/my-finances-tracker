import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

// The app opens on the cashflow ledger (DESIGN_PLAN §3.5: `/cashflow` is also `/`).
export const load: PageLoad = () => {
	redirect(307, '/cashflow');
};
