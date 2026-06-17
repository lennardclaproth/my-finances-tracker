/** Offset pagination metadata returned by list endpoints (`pagination` envelope). */
export interface Pagination {
	limit: number;
	offset: number;
	count: number;
	total: number;
}

/** A list response wrapped in the standard `{ pagination, data }` envelope. */
export interface PaginatedResponse<T> {
	pagination: Pagination;
	data: T[];
}
