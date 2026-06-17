import { apiGet } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import type {
	AssetClassDetails,
	AssetClassesQuery,
	AssetClassesResponse,
	AssetSnapshotsQuery,
	AssetSnapshotsResponse
} from '$lib/api/types';
import { assetClassDetails, assetClasses, assetSnapshots } from '$lib/data/fixtures/assets';
import { clone, delay } from './_mock';

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
