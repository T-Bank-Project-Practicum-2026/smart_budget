const BASE_URL = "/api";

export type ApiResult<T = any> =
  | {
      success: true;
      status: number;
      response: Response;
      data: string;
      json: T;
    }
  | {
      success: false;
      status: number;
      response: Response | null;
      error: string;
      json: any | null;
    };

export type Method = "POST" | "GET";

async function apiRequest<T = any>(
  method: Method,
  endpoint: string,
  options: RequestInit = {},
): Promise<ApiResult<T>> {
  try {
    const response = await fetch(`${BASE_URL}${endpoint}`, {
      method,
      ...options,
      credentials: "include",
      headers: {
        ...options.headers,
        "X-CSRF-Protection": "1",
      },
    });

    const text = await response.text();

    let parsedJson: any = null;
    try {
      if (text) parsedJson = JSON.parse(text);
    } catch {
      parsedJson = null;
    }

    if (!response.ok) {
      return {
        success: false,
        status: response.status,
        response,
        error: text || `Ошибка со статусом ${response.status}`,
        json: parsedJson,
      };
    }

    return {
      success: true,
      status: response.status,
      data: text,
      response,
      json: parsedJson as T,
    };
  } catch (networkError: any) {
    return {
      success: false,
      status: 0,
      response: null,
      error: networkError?.message || "Сетевая ошибка",
      json: null,
    };
  }
}

export async function POST<T = any>(
  endpoint: string,
  options: RequestInit = {},
): Promise<ApiResult<T>> {
  return apiRequest<T>("POST", endpoint, options);
}

export async function GET<T = any>(
  endpoint: string,
  options: RequestInit = {},
): Promise<ApiResult<T>> {
  return apiRequest<T>("GET", endpoint, options);
}
