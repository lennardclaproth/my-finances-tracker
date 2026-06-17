import { browser } from '$app/environment';
import { apiBaseUrl, useMocks } from '$lib/api/config';

/**
 * Realtime refresh (DESIGN_PLAN §5.7). Subscribes to the account WebSocket (`/ws/accounts/{id}`) and
 * debounces an `onRefresh` callback (~250ms) when relevant backend events arrive. No-ops on the server
 * and in mock mode, so the portal runs standalone without a WS backend.
 */
/** Known events: `assets.rebuilt`, `import.completed`, `bulk_tag.completed`, `portfolio.rebuilt`. */
export type RealtimeEvent = string;

export interface RealtimeOptions {
	accountId: string;
	onRefresh: () => void;
	onEvent?: (event: string, payload: unknown) => void;
	/** Only react to these event types (defaults to all). */
	events?: RealtimeEvent[];
	debounceMs?: number;
}

export interface RealtimeConnection {
	disconnect: () => void;
}

function wsBaseUrl(): string {
	return apiBaseUrl.replace(/^http/, 'ws');
}

export function connectRealtime(options: RealtimeOptions): RealtimeConnection {
	// Standalone / SSR: nothing to connect to.
	if (!browser || useMocks || !apiBaseUrl) {
		return { disconnect() {} };
	}

	const debounceMs = options.debounceMs ?? 250;
	let timer: ReturnType<typeof setTimeout> | undefined;
	const socket = new WebSocket(`${wsBaseUrl()}/ws/accounts/${options.accountId}`);

	socket.addEventListener('message', (message: MessageEvent) => {
		let parsed: { type?: string; event?: string };
		try {
			parsed = JSON.parse(message.data) as { type?: string; event?: string };
		} catch {
			return;
		}
		const event = parsed.type ?? parsed.event ?? '';
		if (options.events && !options.events.includes(event)) return;
		options.onEvent?.(event, parsed);
		if (timer) clearTimeout(timer);
		timer = setTimeout(() => options.onRefresh(), debounceMs);
	});

	return {
		disconnect() {
			if (timer) clearTimeout(timer);
			try {
				socket.close();
			} catch {
				// already closing/closed
			}
		}
	};
}
