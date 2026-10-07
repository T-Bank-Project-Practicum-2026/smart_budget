import assert from "node:assert/strict";
import { test } from "node:test";
import { setAccount } from "../src/features/auth/coordination.ts";
import {
  disablePush,
  isPushOptedOut,
} from "../src/features/push/subscription.ts";

const storage = new Map<string, string>();
Object.defineProperty(globalThis, "localStorage", {
  configurable: true,
  value: {
    getItem: (key: string) => storage.get(key) ?? null,
    setItem: (key: string, value: string) => storage.set(key, value),
  },
});

test("successful browser unsubscribe is reflected even if server fails; deletion retries without local subscription", async () => {
  storage.clear();
  setAccount("a");
  let exists = true;
  const sub = {
    endpoint: "https://fcm.googleapis.com/test",
    unsubscribe: async () => {
      exists = false;
      return true;
    },
  };
  Object.defineProperty(globalThis, "navigator", {
    configurable: true,
    value: {
      serviceWorker: {
        getRegistration: async () => ({
          pushManager: { getSubscription: async () => (exists ? sub : null) },
        }),
      },
    },
  });
  globalThis.fetch = async () =>
    Response.json({ error: "service_unavailable" }, { status: 503 });
  assert.deepEqual(await disablePush(), {
    browserDisabled: true,
    serverDisabled: false,
  });
  assert.equal(isPushOptedOut(), true);
  let removals = 0;
  globalThis.fetch = async (_input, init) => {
    assert.equal(init?.method, "DELETE");
    removals++;
    return new Response(null, { status: 204 });
  };
  assert.deepEqual(await disablePush(), {
    browserDisabled: true,
    serverDisabled: true,
  });
  assert.equal(removals, 1);
});

test("server deletion still completes if browser unsubscribe rejects", async () => {
  storage.clear();
  setAccount("a");
  const sub = {
    endpoint: "https://fcm.googleapis.com/test",
    unsubscribe: async () => {
      throw new Error("browser refused");
    },
  };
  Object.defineProperty(globalThis, "navigator", {
    configurable: true,
    value: {
      serviceWorker: {
        getRegistration: async () => ({
          pushManager: { getSubscription: async () => sub },
        }),
      },
    },
  });
  globalThis.fetch = async () => new Response(null, { status: 204 });
  assert.deepEqual(await disablePush(), {
    browserDisabled: false,
    serverDisabled: true,
  });
});
