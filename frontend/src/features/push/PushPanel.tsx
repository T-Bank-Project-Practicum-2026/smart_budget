import { useEffect, useRef, useState } from "react";
import { ApiError, errorMessage } from "../../api";
import { authenticatedRequest as request } from "../auth/api";
import {
  disablePush,
  flushPushRemovals,
  hasPushSubscription,
  isPushOptedOut,
  publishPushChange,
  setPushOptedOut,
  watchPush,
} from "./subscription";
import { accountSnapshot } from "../auth/coordination";

type Notice = { id: string; title: string; body: string; createdAt: string };
const supported =
  window.isSecureContext &&
  "serviceWorker" in navigator &&
  "PushManager" in window &&
  "Notification" in window;
function decodeKey(key: string): Uint8Array<ArrayBuffer> {
  const value = atob(key.replace(/-/g, "+").replace(/_/g, "/"));
  return Uint8Array.from(value, (char) => char.charCodeAt(0));
}

export default function PushPanel({
  disabled = false,
}: {
  disabled?: boolean;
}) {
  const [ready, setReady] = useState(false);
  const [subscribed, setSubscribed] = useState(false);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [notices, setNotices] = useState<Notice[]>([]);
  const requestKey = useRef<string | null>(null);
  const registration = useRef<ServiceWorkerRegistration | null>(null);
  const key = useRef("");
  const alive = useRef(true);

  useEffect(() => {
    alive.current = true;
    let cancelled = false;
    const current = accountSnapshot();
    const valid = () =>
      !cancelled && accountSnapshot().revision === current.revision;
    async function syncState() {
      if (!supported) return;
      const exists = await hasPushSubscription();
      if (valid()) setSubscribed(exists);
    }
    const stopWatching = watchPush(() => {
      void syncState().catch(() => {});
    });
    async function init() {
      if (!supported) return;
      await flushPushRemovals();
      if (!valid()) return;
      await navigator.serviceWorker.register("/sw.js");
      const reg = await navigator.serviceWorker.ready;
      if (!valid()) return;
      registration.current = reg;
      const sub = await reg.pushManager.getSubscription();
      if (!valid()) return;
      if (sub && Notification.permission === "granted" && !isPushOptedOut()) {
        await request("/push/subscription", {
          method: "PUT",
          body: sub.toJSON(),
        });
      }
      await syncState();
    }
    void request<Notice[]>("/notifications")
      .then((list) => {
        if (valid()) setNotices(list);
      })
      .catch((error: unknown) => {
        if (valid()) setMessage(errorMessage(error));
      });
    void init()
      .catch((error: unknown) => {
        if (valid()) setMessage(errorMessage(error));
      })
      .finally(() => {
        if (valid()) setReady(true);
      });
    return () => {
      cancelled = true;
      alive.current = false;
      stopWatching();
    };
  }, []);

  async function enable() {
    setBusy(true);
    setMessage("");
    const snapshot = accountSnapshot();
    const check = () => {
      if (!alive.current || accountSnapshot().revision !== snapshot.revision)
        throw new ApiError(409, "account_changed");
    };
    try {
      // Request permission directly from the button gesture.
      const permission = await Notification.requestPermission();
      check();
      if (permission !== "granted")
        throw new Error(
          "Уведомления запрещены. Разрешите их в настройках сайта.",
        );
      if (!(await flushPushRemovals()))
        throw new Error(
          "Не удалось завершить прежнее отключение push. Повторите попытку.",
        );
      check();
      key.current = (
        await request<{ publicKey: string }>("/push/key")
      ).publicKey;
      check();
      await navigator.serviceWorker.register("/sw.js");
      const reg = await navigator.serviceWorker.ready;
      check();
      registration.current = reg;
      let sub = await reg.pushManager.getSubscription();
      check();
      const subscribe = () =>
        reg.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: decodeKey(key.current),
        });
      sub ??= await subscribe();
      check();
      try {
        await request("/push/subscription", {
          method: "PUT",
          body: sub.toJSON(),
        });
      } catch (error) {
        // Never transfer another account's endpoint. Retire it and obtain a new one.
        if (
          !(error instanceof ApiError) ||
          error.code !== "subscription_conflict"
        )
          throw error;
        if (!(await sub.unsubscribe()))
          throw new Error("Не удалось отключить прежнюю подписку браузера.");
        check();
        sub = await subscribe();
        check();
        await request("/push/subscription", {
          method: "PUT",
          body: sub.toJSON(),
        });
      }
      check();
      setPushOptedOut(false);
      setSubscribed(true);
      setMessage("Уведомления включены.");
      publishPushChange();
    } catch (error) {
      if (alive.current) setMessage(errorMessage(error));
    } finally {
      if (alive.current) setBusy(false);
    }
  }
  async function disable() {
    setBusy(true);
    setMessage("");
    try {
      const result = await disablePush();
      if (!alive.current) return;
      setSubscribed(!result.browserDisabled);
      setMessage(
        result.browserDisabled
          ? result.serverDisabled
            ? "Уведомления выключены."
            : "Push в браузере выключены. Удаление на сервере будет повторено при следующем подключении."
          : "Не удалось отключить подписку браузера. Повторите попытку.",
      );
    } catch (error) {
      if (alive.current) {
        setSubscribed(await hasPushSubscription());
        setMessage(errorMessage(error));
      }
    } finally {
      if (alive.current) setBusy(false);
    }
  }
  async function send() {
    setBusy(true);
    setMessage("");
    try {
      requestKey.current ??= crypto.randomUUID();
      await request("/notifications/test", {
        method: "POST",
        headers: {
          "Idempotency-Key": requestKey.current,
        },
      });
      requestKey.current = null;
      setMessage(
        "Уведомление поставлено в очередь. Дождитесь сообщения браузера.",
      );
      setNotices(await request<Notice[]>("/notifications"));
    } catch (error) {
      setMessage(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  return (
    <section>
      <h2>Уведомления</h2>
      <p>Включите push и отправьте себе тестовое сообщение.</p>
      {!supported ? (
        <p>
          Web Push недоступен в этом браузере. Откройте сайт через HTTPS или
          localhost.
        </p>
      ) : (
        <div className="actions">
          <button
            disabled={!ready || busy || disabled}
            onClick={() => void (subscribed ? disable() : enable())}
          >
            {subscribed ? "Выключить push" : "Включить push"}
          </button>
          <button
            disabled={
              !ready || busy || disabled || !subscribed || isPushOptedOut()
            }
            onClick={() => void send()}
          >
            Отправить тест
          </button>
        </div>
      )}
      <p role="status">{!ready ? "Подключаемся…" : message}</p>
      {notices.length > 0 && (
        <>
          <h2>История</h2>
          <ul className="notices">
            {notices.map((notice) => (
              <li key={notice.id}>
                <strong>{notice.title}</strong>
                <p>{notice.body}</p>
                <time>{new Date(notice.createdAt).toLocaleString()}</time>
              </li>
            ))}
          </ul>
        </>
      )}
    </section>
  );
}
