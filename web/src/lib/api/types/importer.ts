/**
 * Import types. The import endpoints accept `multipart/form-data` uploads (a file plus a few
 * form fields) and return `202 Accepted` with the durable import record's id + initial status.
 */

/** `POST /imports/*` — mirrors `importer.ImportAcceptedResponse`. */
export interface ImportAcceptedResponse {
	import_id: string;
	status: string;
}

/** Form fields for `POST /imports/cashflow`. */
export interface CashflowImportInput {
	file: File;
	vendor_id: string;
	account_id: string;
}

/** Form fields for `POST /imports/portfolio`. */
export interface PortfolioImportInput {
	file: File;
	vendor_id: string;
	account_id: string;
}

/** Form fields for `POST /imports/eod`. */
export interface EODImportInput {
	file: File;
	listing_id: string;
}
