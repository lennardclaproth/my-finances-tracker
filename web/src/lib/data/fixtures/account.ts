import { DEMO_ACCOUNT_ID } from '$lib/api/config';
import type { Account } from '$lib/api/types';

/** The single demo account the fixtures are scoped to. */
export const demoAccount: Account = {
	id: DEMO_ACCOUNT_ID,
	name: 'Demo account'
};
