/** Minimal account identity. Mirrors `account.AccountResponse`. */
export interface Account {
	id: string;
	name: string;
	external_id?: string | null;
}

/** `GET /accounts` returns a bare array of accounts. */
export type AccountsResponse = Account[];

/** `POST /accounts` request — mirrors `account.CreateAccountRequest`. */
export interface CreateAccountRequest {
	name: string;
	external_id?: string;
}

/** `POST /accounts` — mirrors `account.CreateAccountResponse`. */
export interface CreateAccountResponse {
	id: string;
}
