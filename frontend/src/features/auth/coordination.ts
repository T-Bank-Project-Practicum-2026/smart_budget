type Listener = () => void;
const listeners = new Set<Listener>();
const storageKey = "budget-auth-change";
let accountID: string | null = null;
let revision = 0;
const channel =
  typeof window !== "undefined" && "BroadcastChannel" in window
    ? new BroadcastChannel("budget-auth")
    : null;

export function accountSnapshot() {
  return { id: accountID, revision };
}

export function setAccount(id: string | null) {
  if (accountID !== id) revision++;
  accountID = id;
}

export function invalidateSession() {
  accountID = null;
  revision++;
  for (const listener of listeners) listener();
}

export function publishSessionChange() {
  invalidateSession();
  channel?.postMessage("changed");
  // Storage events cover browsers without BroadcastChannel. Never store tokens.
  if (!channel) {
    try {
      localStorage.setItem(storageKey, crypto.randomUUID());
    } catch {
      /* Storage may be unavailable. */
    }
  }
}

if (channel) channel.onmessage = () => invalidateSession();
if (typeof window !== "undefined") {
  window.addEventListener("storage", (event) => {
    if (event.key === storageKey) invalidateSession();
  });
}

export function watchSession(listener: Listener) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export async function withAuthLock<T>(action: () => Promise<T>): Promise<T> {
  if (typeof navigator !== "undefined" && navigator.locks) {
    return navigator.locks.request("budget-auth-cookie", action);
  }
  return action();
}
