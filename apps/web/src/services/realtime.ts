export type DataChangedEvent = "portfolio.rebuilt" | "import.completed" | "bulk_tag.completed" | "assets.rebuilt";

export interface DataChangedMessage {
  type: "data_changed";
  event: DataChangedEvent;
  account_id: string;
  timestamp: string;
}

type Listener = (message: DataChangedMessage) => void;

const rawApiBase = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.trim();
const API_BASE = rawApiBase && rawApiBase !== "/" ? rawApiBase : "/api";

function normalizePath(path: string): string {
  return path.startsWith("/") ? path : `/${path}`;
}

function toWebSocketURL(path: string): string {
  const normalizedPath = normalizePath(path);

  if (API_BASE.startsWith("ws://") || API_BASE.startsWith("wss://")) {
    const base = API_BASE.endsWith("/") ? API_BASE.slice(0, -1) : API_BASE;
    return `${base}${normalizedPath}`;
  }

  const normalizedBase = API_BASE.endsWith("/") ? API_BASE.slice(0, -1) : API_BASE;
  let httpURL = "";

  if (normalizedBase.startsWith("http://") || normalizedBase.startsWith("https://")) {
    httpURL = `${normalizedBase}${normalizedPath}`;
  } else {
    const pathBase = normalizePath(normalizedBase);
    httpURL = `${window.location.origin}${pathBase}${normalizedPath}`;
  }

  if (httpURL.startsWith("https://")) {
    return `wss://${httpURL.slice("https://".length)}`;
  }
  if (httpURL.startsWith("http://")) {
    return `ws://${httpURL.slice("http://".length)}`;
  }
  return httpURL;
}

function reportRealtimeError(message: string, details?: unknown): void {
  if (typeof console === "undefined") {
    return;
  }
  if (details === undefined) {
    console.warn(`[realtime] ${message}`);
    return;
  }
  console.warn(`[realtime] ${message}`, details);
}

class RealtimeClient {
  private socket: WebSocket | null = null;
  private accountId = "";
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectAttempts = 0;
  private listeners = new Set<Listener>();
  private awaitUserInteractionReconnect = false;

  public constructor() {
    if (typeof window === "undefined") {
      return;
    }
    window.addEventListener("click", this.handleUserInteraction, { passive: true });
    window.addEventListener("pointerdown", this.handleUserInteraction, { passive: true });
  }

  public setAccountId(accountId: string): void {
    const next = accountId.trim();
    if (next === this.accountId) {
      return;
    }

    this.accountId = next;
    this.awaitUserInteractionReconnect = false;
    this.clearReconnectTimer();
    this.reconnectAttempts = 0;
    this.disconnect();

    if (this.accountId !== "") {
      this.connect();
    }
  }

  public subscribe(listener: Listener): () => void {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  private connect(): void {
    if (this.accountId === "") {
      return;
    }
    if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
      return;
    }

    const wsURL = toWebSocketURL(`/ws/accounts/${encodeURIComponent(this.accountId)}`);
    const socket = new WebSocket(wsURL);
    this.socket = socket;

    socket.onopen = () => {
      if (this.socket !== socket) {
        return;
      }
      this.reconnectAttempts = 0;
      this.awaitUserInteractionReconnect = false;
    };

    socket.onmessage = (event) => {
      const message = this.parseMessage(event.data);
      if (!message) {
        return;
      }
      for (const listener of this.listeners) {
        listener(message);
      }
    };

    socket.onerror = (event) => {
      reportRealtimeError("websocket connection error", event);
    };

    socket.onclose = (event) => {
      if (this.socket !== socket) {
        return;
      }
      this.socket = null;
      if (this.accountId === "") {
        return;
      }
      if (event.code === 4001) {
        this.awaitUserInteractionReconnect = true;
        this.clearReconnectTimer();
        return;
      }
      if (event.code !== 1000) {
        reportRealtimeError("websocket closed", { code: event.code, reason: event.reason });
      }
      this.awaitUserInteractionReconnect = false;
      this.scheduleReconnect();
    };
  }

  private disconnect(): void {
    if (!this.socket) {
      return;
    }

    const current = this.socket;
    this.socket = null;
    current.onclose = null;
    current.onmessage = null;
    current.onerror = null;
    current.onopen = null;
    current.close();
  }

  private scheduleReconnect(): void {
    if (typeof window === "undefined") {
      return;
    }
    if (this.reconnectTimer || this.awaitUserInteractionReconnect || this.accountId === "") {
      return;
    }

    const delay = this.nextBackoff();
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null;
      this.reconnectAttempts += 1;
      this.connect();
    }, delay);
  }

  private nextBackoff(): number {
    const steps = [500, 1000, 2000, 5000, 10000];
    const index = Math.min(this.reconnectAttempts, steps.length - 1);
    return steps[index] ?? 10000;
  }

  private clearReconnectTimer(): void {
    if (!this.reconnectTimer) {
      return;
    }
    clearTimeout(this.reconnectTimer);
    this.reconnectTimer = null;
  }

  private parseMessage(raw: unknown): DataChangedMessage | null {
    if (typeof raw !== "string") {
      reportRealtimeError("ignoring websocket message with non-string payload");
      return null;
    }
    try {
      const parsed = JSON.parse(raw) as Partial<DataChangedMessage>;
      if (
        parsed.type !== "data_changed" ||
        typeof parsed.event !== "string" ||
        typeof parsed.account_id !== "string" ||
        typeof parsed.timestamp !== "string"
      ) {
        reportRealtimeError("ignoring websocket message with invalid shape", parsed);
        return null;
      }
      if (
        parsed.event !== "portfolio.rebuilt" &&
        parsed.event !== "import.completed" &&
        parsed.event !== "bulk_tag.completed" &&
        parsed.event !== "assets.rebuilt"
      ) {
        reportRealtimeError("ignoring websocket message with unknown event", parsed.event);
        return null;
      }
      return parsed as DataChangedMessage;
    } catch (error) {
      reportRealtimeError("failed parsing websocket message", error);
      return null;
    }
  }

  private handleUserInteraction = (): void => {
    if (!this.awaitUserInteractionReconnect || this.accountId === "") {
      return;
    }
    if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
      return;
    }

    this.awaitUserInteractionReconnect = false;
    this.clearReconnectTimer();
    this.reconnectAttempts = 0;
    this.connect();
  };
}

const realtimeClient = new RealtimeClient();

export function getRealtimeClient(): RealtimeClient {
  return realtimeClient;
}
