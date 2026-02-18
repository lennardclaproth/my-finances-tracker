export class ApiError extends Error {
  public readonly status: number;

  public constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

type QueryValue = string | number | boolean | undefined;
const rawApiBase = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.trim();
const API_BASE = rawApiBase && rawApiBase !== "/" ? rawApiBase : "/api";

interface JsonRequestOptions extends RequestInit {
  query?: Record<string, QueryValue>;
}

interface FormRequestOptions extends Omit<RequestInit, "body"> {
  body: FormData;
  query?: Record<string, QueryValue>;
}

function buildQueryString(query?: Record<string, QueryValue>): string {
  if (!query) {
    return "";
  }

  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value === undefined) {
      continue;
    }
    params.set(key, String(value));
  }
  const queryString = params.toString();
  return queryString ? `?${queryString}` : "";
}

function normalizePath(path: string): string {
  return path.startsWith("/") ? path : `/${path}`;
}

function joinUrl(base: string, path: string): string {
  const normalizedBase = base.endsWith("/") ? base.slice(0, -1) : base;
  return `${normalizedBase}${normalizePath(path)}`;
}

function extractErrorMessage(payload: unknown, status: number): string {
  if (typeof payload === "string" && payload.trim() !== "") {
    return payload;
  }
  if (payload && typeof payload === "object") {
    const values = Object.values(payload as Record<string, unknown>)
      .map((value) => String(value))
      .filter(Boolean);
    if (values.length > 0) {
      return values.join(", ");
    }
  }
  return `Request failed with status ${status}`;
}

export async function requestJson<T>(path: string, options: JsonRequestOptions = {}): Promise<T> {
  const { query, headers, ...rest } = options;
  const requestPath = `${joinUrl(API_BASE, path)}${buildQueryString(query)}`;

  const response = await fetch(requestPath, {
    ...rest,
    headers: {
      ...(headers ?? {}),
      ...(rest.body ? { "Content-Type": "application/json" } : {}),
    },
  });

  const contentType = response.headers.get("content-type") ?? "";
  const hasJsonBody = contentType.includes("application/json");
  const payload = hasJsonBody ? await response.json() : await response.text();

  if (!response.ok) {
    throw new ApiError(response.status, extractErrorMessage(payload, response.status));
  }

  return payload as T;
}

export async function requestFormData<T>(path: string, options: FormRequestOptions): Promise<T> {
  const { query, headers, body, ...rest } = options;
  const requestPath = `${joinUrl(API_BASE, path)}${buildQueryString(query)}`;

  const response = await fetch(requestPath, {
    ...rest,
    body,
    headers: {
      ...(headers ?? {}),
    },
  });

  const contentType = response.headers.get("content-type") ?? "";
  const hasJsonBody = contentType.includes("application/json");
  const payload = hasJsonBody ? await response.json() : await response.text();

  if (!response.ok) {
    throw new ApiError(response.status, extractErrorMessage(payload, response.status));
  }

  return payload as T;
}
