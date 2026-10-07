import { ApiError, request, type RequestOptions } from "../../api.ts";
import {
  accountSnapshot,
  invalidateSession,
  publishSessionChange,
  withAuthLock,
} from "./coordination.ts";

export type User = {
  id: string;
  email: string;
  timezone: string;
  createdAt: string;
};
export type Credentials = { email: string; password: string };
export type Registration = Credentials & { timezone: string };
let refreshing: Promise<void> | undefined;

async function refreshSession() {
  await withAuthLock(async () => {
    // Another tab may already have refreshed while this tab waited for the lock.
    try {
      await request<User>("/auth/me");
      return;
    } catch (error) {
      if (!(error instanceof ApiError) || error.status !== 401) throw error;
    }
    try {
      await request<void>("/auth/refresh", { method: "POST" });
    } catch (error) {
      if (!(error instanceof ApiError) || error.status !== 401) throw error;
      // Compatibility fallback when Web Locks is unavailable: the winner's
      // Set-Cookie may arrive just after the losing refresh response.
      for (let attempt = 0; attempt < 3; attempt++) {
        await new Promise((resolve) => setTimeout(resolve, 100));
        try {
          await request<User>("/auth/me");
          return;
        } catch (probe) {
          if (!(probe instanceof ApiError) || probe.status !== 401) throw probe;
        }
      }
      throw error;
    }
  });
}

export async function authenticatedRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const snapshot = accountSnapshot();
  const discovery = path === "/auth/me";
  if (!discovery && !snapshot.id) throw new ApiError(409, "account_changed");
  const guarded = discovery
    ? options
    : {
        ...options,
        headers: { ...options.headers, "X-Account-ID": snapshot.id! },
      };
  const checkAccount = () => {
    if (!discovery && accountSnapshot().revision !== snapshot.revision)
      throw new ApiError(409, "account_changed");
  };
  try {
    let result: T;
    try {
      result = await request<T>(path, guarded);
    } catch (error) {
      if (!(error instanceof ApiError) || error.status !== 401) throw error;
      refreshing ??= refreshSession().finally(() => {
        refreshing = undefined;
      });
      await refreshing;
      checkAccount();
      result = await request<T>(path, guarded);
    }
    checkAccount();
    return result;
  } catch (error) {
    if (
      error instanceof ApiError &&
      (error.status === 401 || error.code === "account_changed") &&
      accountSnapshot().id &&
      accountSnapshot().revision === snapshot.revision
    )
      invalidateSession();
    throw error;
  }
}

async function credentialsRequest(
  path: string,
  body: Credentials | Registration,
) {
  return withAuthLock(async () => {
    const user = await request<User>(path, { method: "POST", body });
    publishSessionChange();
    return user;
  });
}

export const authApi = {
  register: (body: Registration) => credentialsRequest("/auth/register", body),
  login: (body: Credentials) => credentialsRequest("/auth/login", body),
  logout: () => {
    const expected = accountSnapshot().id;
    return withAuthLock(async () => {
      // Do not log out an account which another tab has just selected.
      if (expected !== accountSnapshot().id)
        throw new ApiError(409, "account_changed");
      await request<void>("/auth/logout", {
        method: "POST",
        headers: expected ? { "X-Account-ID": expected } : {},
      });
      publishSessionChange();
    });
  },
  me: () => authenticatedRequest<User>("/auth/me"),
};
