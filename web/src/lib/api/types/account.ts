/** Minimal account identity. The portal is single-account for now; the API exposes account by id. */
export interface Account {
	id: string;
	name: string;
}

/** `POST /accounts` — mirrors `account.CreateAccountResponse`. */
export interface CreateAccountResponse {
	id: string;
}
