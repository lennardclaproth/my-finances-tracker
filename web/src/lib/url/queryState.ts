import { goto } from '$app/navigation';

/**
 * SvelteKit navigation wrapper for URL-as-state (DESIGN_PLAN §5.8). Pages compute the next
 * URLSearchParams (via routeQuery's serializeQuery) and push them here; tables/footers stay
 * presentational. Uses keepFocus + noScroll so filtering doesn't jump the page.
 */
export function pushQuery(
	pathname: string,
	params: URLSearchParams,
	opts: { replace?: boolean } = {}
): Promise<void> {
	const qs = params.toString();
	const url = qs ? `${pathname}?${qs}` : pathname;
	// `pathname` is the caller's already-correct route; this is a same-app query-string update.
	// eslint-disable-next-line svelte/no-navigation-without-resolve
	return goto(url, { keepFocus: true, noScroll: true, replaceState: opts.replace ?? false });
}
