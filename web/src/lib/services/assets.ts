import { apiGet, apiSend } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import type {
	AdjustAssetWorthRequest,
	AssetClassDetails,
	AssetClassesQuery,
	AssetClassesResponse,
	AssetSnapshotsQuery,
	AssetSnapshotsResponse,
	CreateAssetClassRequest,
	CreateAssetClassResponse,
	CreateAssetRequest,
	CreateAssetResponse,
	DeleteAssetClassRequest,
	SetAssetWorthRequest,
	UpdateAssetClassRequest
} from '$lib/api/types';
import { assetClassDetails, assetClasses, assetSnapshots } from '$lib/data/fixtures/assets';
import { clone, delay, mockId } from './_mock';

/** `GET /assets/classes` */
export async function listAssetClasses(query: AssetClassesQuery): Promise<AssetClassesResponse> {
	if (useMocks) {
		await delay();
		return clone(assetClasses.filter((c) => (query.include_archived ? true : !c.archived)));
	}
	return apiGet<AssetClassesResponse>('/assets/classes', { ...query });
}

/** `GET /assets/classes/{class_id}` */
export async function getAssetClassDetails(
	classId: string,
	accountId: string
): Promise<AssetClassDetails> {
	if (useMocks) {
		await delay();
		const details = assetClassDetails[classId];
		if (details) return clone(details);
		// Fall back to a details view synthesized from the summary for classes without rich fixtures.
		const summary = assetClasses.find((c) => c.id === classId) ?? assetClasses[0];
		return clone({ class: summary, assets: [], growth: [], mutations: [] });
	}
	return apiGet<AssetClassDetails>(`/assets/classes/${classId}`, { account_id: accountId });
}

/** `GET /assets/snapshots` */
export async function getAssetSnapshots(
	query: AssetSnapshotsQuery
): Promise<AssetSnapshotsResponse> {
	if (useMocks) {
		await delay();
		return clone(
			assetSnapshots.filter(
				(p) => (!query.from || p.date >= query.from) && (!query.to || p.date <= query.to)
			)
		);
	}
	return apiGet<AssetSnapshotsResponse>('/assets/snapshots', { ...query });
}

/** `POST /assets/classes` */
export async function createAssetClass(
	body: CreateAssetClassRequest
): Promise<CreateAssetClassResponse> {
	if (useMocks) {
		await delay();
		return { id: mockId() };
	}
	return apiSend<CreateAssetClassResponse>('POST', '/assets/classes', body);
}

/** `PATCH /assets/classes` */
export async function updateAssetClass(body: UpdateAssetClassRequest): Promise<void> {
	if (useMocks) {
		await delay();
		return;
	}
	await apiSend<unknown>('PATCH', '/assets/classes', body);
}

/** `DELETE /assets/classes/{class_id}` */
export async function deleteAssetClass(
	classId: string,
	body: DeleteAssetClassRequest
): Promise<void> {
	if (useMocks) {
		await delay();
		return;
	}
	await apiSend<unknown>('DELETE', `/assets/classes/${classId}`, body);
}

/** `POST /assets` */
export async function createAsset(body: CreateAssetRequest): Promise<CreateAssetResponse> {
	if (useMocks) {
		await delay();
		return { id: mockId(), name: body.name };
	}
	return apiSend<CreateAssetResponse>('POST', '/assets', body);
}

/** `PUT /assets/{asset_id}/worth` */
export async function setAssetWorth(assetId: string, body: SetAssetWorthRequest): Promise<void> {
	if (useMocks) {
		await delay();
		return;
	}
	await apiSend<unknown>('PUT', `/assets/${assetId}/worth`, body);
}

/** `PUT /assets/{asset_id}/adjust` */
export async function adjustAssetWorth(
	assetId: string,
	body: AdjustAssetWorthRequest
): Promise<void> {
	if (useMocks) {
		await delay();
		return;
	}
	await apiSend<unknown>('PUT', `/assets/${assetId}/adjust`, body);
}
