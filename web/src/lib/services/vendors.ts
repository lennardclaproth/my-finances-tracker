import { apiGet } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import type { VendorsResponse } from '$lib/api/types';
import { vendors } from '$lib/data/fixtures/vendors';
import { clone, delay } from './_mock';

/** `GET /vendors` */
export async function listVendors(): Promise<VendorsResponse> {
	if (useMocks) {
		await delay();
		return clone(vendors.filter((v) => v.active));
	}
	return apiGet<VendorsResponse>('/vendors');
}
