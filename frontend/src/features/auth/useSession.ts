import { useEffect, useRef, useState } from "react";
import { ApiError, errorMessage } from "../../api";
import { disablePush } from "../push/subscription";
import { authApi, type User } from "./api";
import { accountSnapshot, setAccount, watchSession } from "./coordination";

export function useSession() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const sequence = useRef(0);

  useEffect(() => {
    let disposed = false;
    async function synchronize(clear = false) {
      const ticket = ++sequence.current;
      if (clear) {
        setLoading(true);
        setUser(null);
      }
      try {
        const value = await authApi.me();
        if (disposed || ticket !== sequence.current) return;
        setAccount(value.id);
        setUser(value);
        setMessage("");
      } catch (error) {
        if (disposed || ticket !== sequence.current) return;
        setAccount(null);
        setUser(null);
        if (!(error instanceof ApiError && error.status === 401))
          setMessage(errorMessage(error));
      } finally {
        if (!disposed && ticket === sequence.current) setLoading(false);
      }
    }
    const refresh = () => {
      void synchronize();
    };
    const visible = () => {
      if (document.visibilityState === "visible") refresh();
    };
    const unsubscribe = watchSession(() => {
      void synchronize(true);
    });
    window.addEventListener("focus", refresh);
    document.addEventListener("visibilitychange", visible);
    refresh();
    return () => {
      disposed = true;
      sequence.current++;
      unsubscribe();
      window.removeEventListener("focus", refresh);
      document.removeEventListener("visibilitychange", visible);
    };
  }, []);

  function acceptUser() {
    // The session-change listener loads the current cookie owner. A delayed
    // login response must not overwrite a newer login from another tab.
    setMessage("");
  }

  async function logout() {
    setBusy(true);
    setMessage("");
    try {
      const expected = accountSnapshot().id;
      const result = await disablePush().catch(() => null);
      if (expected !== accountSnapshot().id)
        throw new ApiError(409, "account_changed");
      await authApi.logout();
      if (!result?.browserDisabled)
        setMessage(
          "Вы вышли. Новые push отключены сервером; браузерную подписку не удалось удалить.",
        );
    } catch (error) {
      setMessage(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  return { user, loading, busy, message, acceptUser, logout };
}
