const BASE_URL = "/api";

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public code: string | null = null,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

let refreshPromise: Promise<boolean> | null = null;

function refreshSession(): Promise<boolean> {
  refreshPromise ??= fetch(`${BASE_URL}/v1/auth/refresh`, {
    method: "POST",
    credentials: "include",
    headers: { "X-CSRF-Protection": "1" },
  })
    .then((r) => r.ok)
    .catch(() => false)
    .finally(() => {
      refreshPromise = null;
    });
  return refreshPromise;
}

async function request<T>(
  method: "GET" | "POST",
  endpoint: string,
  body?: unknown,
  retried = false,
): Promise<T> {
  const headers = new Headers({
    Accept: "application/json",
    "X-CSRF-Protection": "1",
  });

  let payload: BodyInit | undefined;
  if (body instanceof FormData) {
    payload = body;
  } else if (body !== undefined) {
    headers.set("Content-Type", "application/json");
    payload = JSON.stringify(body);
  }

  let response: Response;
  try {
    response = await fetch(`${BASE_URL}${endpoint}`, {
      method,
      headers,
      body: payload,
      credentials: "include",
    });
  } catch {
    throw new ApiError(0, "Нет соединения с сервером");
  }

  const text = await response.text();
  let json: any = null;
  if (text !== "") {
    try {
      json = JSON.parse(text);
    } catch {
      // не json - текст пригодится для сообщения об ошибке
    }
  }

  // Токен старый
  if (
    response.status === 401 &&
    json?.error === "authentication_required" &&
    !retried &&
    (await refreshSession())
  ) {
    return request<T>(method, endpoint, body, true);
  }

  if (!response.ok) {
    const message = String(
      json?.message ?? json?.error ?? (text || response.statusText),
    );
    throw new ApiError(response.status, message, json?.error ?? null);
  }

  return json as T;
}

export function GET<T>(endpoint: string): Promise<T> {
  return request<T>("GET", endpoint);
}

export function POST<T>(endpoint: string, body?: unknown): Promise<T> {
  return request<T>("POST", endpoint, body);
}
