import { DEMO_ACCOUNT_ID } from '$lib/api/config';
import { listAccounts } from '$lib/services/account';
import type { Account } from '$lib/api/types';

/**
 * Active-account context. The portal is single-account today: on first use it fetches `GET /accounts`
 * and adopts the first account as active, so account-scoped screens (portfolio, assets, realtime)
 * send a real backend id instead of a hard-coded one. In mock mode the demo account is returned.
 */
let accounts = $state<Account[]>([]);
let activeId = $state<string>(DEMO_ACCOUNT_ID);
let loaded = $state(false);
let loadPromise: Promise<void> | null = null;

export const accountStore = {
	get accounts(): Account[] {
		return accounts;
	},
	/** The id every account-scoped request should use. Defaults to the demo id until loaded. */
	get activeId(): string {
		return activeId;
	},
	get loaded(): boolean {
		return loaded;
	},
	/** Fetch the account list once and adopt the first account as active. Idempotent. */
	ensureLoaded(): Promise<void> {
		if (!loadPromise) {
			loadPromise = listAccounts()
				.then((list) => {
					if (list.length > 0) {
						accounts = list;
						activeId = list[0].id;
					}
				})
				.catch(() => {
					// Keep the default id; pages surface data-load failures separately.
				})
				.finally(() => {
					loaded = true;
				});
		}
		return loadPromise;
	}
};
