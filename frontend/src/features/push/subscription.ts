import { ApiError } from "../../api.ts";
import { authenticatedRequest } from "../auth/api.ts";
import { accountSnapshot } from "../auth/coordination.ts";

type PendingRemoval = { userId: string; endpoint: string };
export type DisableResult = {
  browserDisabled: boolean;
  serverDisabled: boolean;
};
const pendingKey = "budget-push-removals";
const changeKey = "budget-push-change";
const preferenceKey = "budget-push-off";
let memoryOptedOut = false;
export function isPushOptedOut() {
  try {
    return localStorage.getItem(preferenceKey) === "1";
  } catch {
    return memoryOptedOut;
  }
}
export function setPushOptedOut(value: boolean) {
  memoryOptedOut = value;
  try {
    localStorage.setItem(preferenceKey, value ? "1" : "0");
  } catch {
    /* Memory fallback. */
  }
}
let memoryPending: PendingRemoval[] = [];
const listeners = new Set<() => void>();
const channel =
  typeof window !== "undefined" && "BroadcastChannel" in window
    ? new BroadcastChannel("budget-push")
    : null;
function changed() {
  for (const listener of listeners) listener();
}
if (channel) channel.onmessage = changed;
if (typeof window !== "undefined")
  window.addEventListener("storage", (event) => {
    if (event.key === changeKey) changed();
  });
export function watchPush(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}
export function publishPushChange() {
  changed();
  channel?.postMessage("changed");
  if (!channel) {
    try {
      localStorage.setItem(changeKey, crypto.randomUUID());
    } catch {
      /* Best effort cross-tab fallback. */
    }
  }
}
function pending(): PendingRemoval[] {
  try {
    const value: unknown = JSON.parse(localStorage.getItem(pendingKey) ?? "[]");
    if (Array.isArray(value))
      return value.filter(
        (v): v is PendingRemoval =>
          typeof v?.userId === "string" && typeof v?.endpoint === "string",
      );
  } catch {
    /* Use memory when storage is unavailable. */
  }
  return memoryPending;
}
function save(value: PendingRemoval[]) {
  memoryPending = value;
  try {
    localStorage.setItem(pendingKey, JSON.stringify(value));
  } catch {
    /* Retried in this tab. */
  }
}
function requireAccount(userId: string | null) {
  if (!userId || accountSnapshot().id !== userId)
    throw new ApiError(409, "account_changed");
}
export async function flushPushRemovals(): Promise<boolean> {
  const userId = accountSnapshot().id;
  requireAccount(userId);
  let succeeded = true;
  for (const item of pending().filter((item) => item.userId === userId)) {
    try {
      requireAccount(userId);
      await authenticatedRequest<void>("/push/subscription", {
        method: "DELETE",
        body: { endpoint: item.endpoint },
      });
      save(
        pending().filter(
          (entry) =>
            entry.userId !== item.userId || entry.endpoint !== item.endpoint,
        ),
      );
    } catch {
      succeeded = false;
    }
  }
  return succeeded;
}
export async function hasPushSubscription(): Promise<boolean> {
  if (!("serviceWorker" in navigator)) return false;
  const registration = await navigator.serviceWorker.getRegistration();
  return !!(await registration?.pushManager?.getSubscription());
}
export async function disablePush(): Promise<DisableResult> {
  const userId = accountSnapshot().id;
  requireAccount(userId);
  setPushOptedOut(true);
  const registration =
    "serviceWorker" in navigator
      ? await navigator.serviceWorker.getRegistration()
      : undefined;
  const subscription = await registration?.pushManager?.getSubscription();
  requireAccount(userId);
  if (subscription) {
    const entries = pending();
    if (
      !entries.some(
        (entry) =>
          entry.userId === userId && entry.endpoint === subscription.endpoint,
      )
    )
      save([...entries, { userId: userId!, endpoint: subscription.endpoint }]);
  }
  const [server] = await Promise.allSettled([
    flushPushRemovals(),
    subscription ? subscription.unsubscribe() : Promise.resolve(true),
  ]);
  // The actual browser state is authoritative even after a rejected unsubscribe.
  const browserDisabled = !(await hasPushSubscription());
  publishPushChange();
  return {
    browserDisabled,
    serverDisabled: server.status === "fulfilled" && server.value,
  };
}
