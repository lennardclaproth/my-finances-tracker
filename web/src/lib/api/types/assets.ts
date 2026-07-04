/**
 * Asset-class summary. Mirrors `assets.ClassResponse`.
 * Money (`current_worth`) is a decimal string. `growth_pct` is a fraction/percent number or null.
 */
export interface AssetClass {
	id: string;
	name: string;
	source: string;
	archived: boolean;
	current_worth: string;
	/** RFC3339 timestamp or null. */
	last_change_at?: string | null;
	growth_pct?: number | null;
	updated_at: string;
}

/** `GET /assets/classes` returns a bare array of asset classes. */
export type AssetClassesResponse = AssetClass[];

/** One asset within a class. Mirrors `assets.AssetResponse`. */
export interface Asset {
	id: string;
	name: string;
	current_worth: string;
	archived: boolean;
	updated_at: string;
}

/** A point on a class's growth timeline. Mirrors `assets.ClassGrowthPointResponse`. */
export interface ClassGrowthPoint {
	/** "YYYY-MM-DD". */
	date: string;
	total_worth: string;
}

/** A worth mutation record. Mirrors `assets.MutationResponse`. */
export interface AssetMutation {
	id: string;
	item_id: string;
	change_type: string;
	direction?: string | null;
	amount: string;
	previous_worth: string;
	new_worth: string;
	class_total_worth: string;
	/** "YYYY-MM-DD". */
	effective_date: string;
	note?: string | null;
	created_at: string;
}

/** `GET /assets/classes/{class_id}` — mirrors `assets.ClassDetailsResponse`. */
export interface AssetClassDetails {
	class: AssetClass;
	assets: Asset[];
	growth: ClassGrowthPoint[];
	mutations: AssetMutation[];
}

/** Account-level total worth snapshot point. Mirrors `assets.SnapshotResponse`. */
export interface AssetSnapshotPoint {
	/** "YYYY-MM-DD". */
	date: string;
	total_worth: string;
}

/** `GET /assets/snapshots` returns a bare array of snapshot points. */
export type AssetSnapshotsResponse = AssetSnapshotPoint[];

/** `POST /assets/classes` request — mirrors `assets.CreateAssetClassRequest`. */
export interface CreateAssetClassRequest {
	account_id: string;
	name: string;
}

/** `POST /assets/classes` — mirrors `assets.CreateAssetClassResponse`. */
export interface CreateAssetClassResponse {
	id: string;
}

/** `PATCH /assets/classes` request — mirrors `assets.UpdateClassRequest`. */
export interface UpdateAssetClassRequest {
	account_id: string;
	id: string;
	name?: string;
	archived?: boolean;
}

/** `DELETE /assets/classes/{class_id}` body — mirrors `assets.DeleteClassRequest`. */
export interface DeleteAssetClassRequest {
	account_id: string;
}

/** `POST /assets` request — mirrors `assets.CreateAssetRequest`. */
export interface CreateAssetRequest {
	account_id: string;
	class_id: string;
	name: string;
	/** Non-negative decimal string. */
	initial_worth: string;
	/** "YYYY-MM-DD". */
	effective_date: string;
	note?: string;
}

/** `POST /assets` — mirrors `assets.CreateAssetResponse`. */
export interface CreateAssetResponse {
	id: string;
	name: string;
}

/** `PUT /assets/{asset_id}/worth` request — mirrors `assets.SetAssetWorthRequest`. */
export interface SetAssetWorthRequest {
	account_id: string;
	/** Non-negative decimal string. */
	worth: string;
	/** "YYYY-MM-DD". */
	effective_date: string;
	note?: string;
}

/** Direction for an asset-worth adjustment. */
export type AssetWorthDirection = 'INCREASE' | 'DECREASE';

/** `PUT /assets/{asset_id}/adjust` request — mirrors `assets.AdjustAssetWorthRequest`. */
export interface AdjustAssetWorthRequest {
	account_id: string;
	direction: AssetWorthDirection;
	/** Non-negative decimal string. */
	amount: string;
	/** "YYYY-MM-DD". */
	effective_date: string;
	note?: string;
}

/** Query filters for `GET /assets/classes`. */
export interface AssetClassesQuery {
	account_id: string;
	include_archived?: boolean;
}

/** Query filters for `GET /assets/snapshots`. */
export interface AssetSnapshotsQuery {
	account_id: string;
	from?: string;
	to?: string;
}
