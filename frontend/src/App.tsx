import { useEffect, useState } from "react";

type Status = "loading" | "ok" | "error";

const messages: Record<Status, string> = {
  loading: "Проверяем подключение…",
  ok: "API работает",
  error: "Не удалось подключиться к API",
};

export default function App() {
  const [status, setStatus] = useState<Status>("loading");
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), 5000);
    let cancelled = false;

    async function checkHealth() {
      try {
        const response = await fetch("/api/health", {
          signal: controller.signal,
          cache: "no-store",
        });
        if (!response.ok) throw new Error("API unavailable");

        const health: unknown = await response.json();
        if (
          typeof health !== "object" ||
          health === null ||
          !("status" in health) ||
          health.status !== "ok"
        ) {
          throw new Error("Unexpected health response");
        }

        if (!cancelled) setStatus("ok");
      } catch {
        if (!cancelled) setStatus("error");
      } finally {
        window.clearTimeout(timeout);
      }
    }

    void checkHealth();
    return () => {
      cancelled = true;
      window.clearTimeout(timeout);
      controller.abort();
    };
  }, [attempt]);

  function retry() {
    setStatus("loading");
    setAttempt((value) => value + 1);
  }

  return (
    <main>
      <h1>Умный бюджет</h1>
      <p className="description">Учёт расходов и накоплений</p>
      <p className={`status ${status}`} role="status">
        {messages[status]}
      </p>
      <button onClick={retry} disabled={status === "loading"}>
        Проверить снова
      </button>
    </main>
  );
}
