import { apiUpload } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import type {
	CashflowImportInput,
	EODImportInput,
	ImportAcceptedResponse,
	PortfolioImportInput
} from '$lib/api/types';
import { delay, mockId } from './_mock';

function mockAccepted(): ImportAcceptedResponse {
	return { import_id: mockId(), status: 'pending' };
}

/** `POST /imports/cashflow` (multipart) */
export async function importCashflow(
	input: CashflowImportInput
): Promise<ImportAcceptedResponse> {
	if (useMocks) {
		await delay();
		return mockAccepted();
	}
	const form = new FormData();
	form.append('file', input.file);
	form.append('vendor_id', input.vendor_id);
	form.append('account_id', input.account_id);
	return apiUpload<ImportAcceptedResponse>('/imports/cashflow', form);
}

/** `POST /imports/portfolio` (multipart) */
export async function importPortfolio(
	input: PortfolioImportInput
): Promise<ImportAcceptedResponse> {
	if (useMocks) {
		await delay();
		return mockAccepted();
	}
	const form = new FormData();
	form.append('file', input.file);
	form.append('vendor_id', input.vendor_id);
	form.append('account_id', input.account_id);
	return apiUpload<ImportAcceptedResponse>('/imports/portfolio', form);
}

/** `POST /imports/eod` (multipart) */
export async function importEOD(input: EODImportInput): Promise<ImportAcceptedResponse> {
	if (useMocks) {
		await delay();
		return mockAccepted();
	}
	const form = new FormData();
	form.append('file', input.file);
	form.append('listing_id', input.listing_id);
	return apiUpload<ImportAcceptedResponse>('/imports/eod', form);
}
