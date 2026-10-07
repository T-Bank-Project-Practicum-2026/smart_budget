const messages: Record<string, string> = {
  invalid_request: "Проверьте поля формы.",
  invalid_email: "Введите корректный email.",
  invalid_password: "Пароль: минимум 8 символов, максимум 72 байта.",
  invalid_timezone: "Неизвестный часовой пояс.",
  account_exists: "Этот email уже зарегистрирован.",
  invalid_credentials: "Неверный email или пароль.",
  authentication_required: "Войдите в аккаунт.",
  account_changed:
    "Аккаунт изменился в другой вкладке. Обновляем страницу аккаунта.",
  too_many_attempts: "Слишком много попыток. Повторите через 15 минут.",
  forbidden_origin: "Этот адрес сайта не разрешён сервером.",
  invalid_subscription: "Не удалось сохранить push-подписку.",
  subscription_conflict:
    "Подписка относится к другому аккаунту. Включите уведомления заново.",
  push_rate_limited: "Подождите 10 секунд перед следующей отправкой.",
};

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string) {
    super(messages[code] ?? "Не удалось выполнить запрос. Попробуйте ещё раз.");
    this.status = status;
    this.code = code;
  }
}

export type RequestOptions = {
  method?: "GET" | "POST" | "PUT" | "DELETE";
  body?: unknown;
  headers?: Record<string, string>;
};

export async function request<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const response = await fetch(`/api/v1${path}`, {
    method: options.method ?? "GET",
    credentials: "same-origin",
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Protection": "1",
      ...options.headers,
    },
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
    signal: AbortSignal.timeout(15000),
  });
  if (!response.ok) {
    const error = (await response.json().catch(() => ({}))) as {
      error?: string;
    };
    throw new ApiError(response.status, error.error ?? "service_unavailable");
  }
  return response.status === 204 ? (undefined as T) : response.json();
}

export function errorMessage(error: unknown): string {
  return error instanceof Error
    ? error.message
    : "Не удалось выполнить действие.";
}
