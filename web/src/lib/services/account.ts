import { apiGet, apiSend } from '$lib/api/client';
import { DEMO_ACCOUNT_ID, useMocks } from '$lib/api/config';
import type {
	Account,
	AccountsResponse,
	CreateAccountRequest,
	CreateAccountResponse
} from '$lib/api/types';
import { delay, mockId } from './_mock';

const demoAccounts: Account[] = [{ id: DEMO_ACCOUNT_ID, name: 'Demo account' }];

/** `GET /accounts` */
export async function listAccounts(): Promise<AccountsResponse> {
	if (useMocks) {
		await delay();
		return demoAccounts.slice();
	}
	return apiGet<AccountsResponse>('/accounts');
}

/** `POST /accounts` */
export async function createAccount(body: CreateAccountRequest): Promise<CreateAccountResponse> {
	if (useMocks) {
		await delay();
		return { id: mockId() };
	}
	return apiSend<CreateAccountResponse>('POST', '/accounts', body);
}
