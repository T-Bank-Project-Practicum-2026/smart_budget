import assert from "node:assert/strict";
import { test } from "node:test";
import { authenticatedRequest, authApi } from "../src/features/auth/api.ts";
import {
  accountSnapshot,
  invalidateSession,
  setAccount,
  watchSession,
} from "../src/features/auth/coordination.ts";

function installLocks() {
  let tail: Promise<unknown> = Promise.resolve();
  Object.defineProperty(globalThis, "navigator", {
    configurable: true,
    value: {
      locks: {
        request: (_name: string, action: () => Promise<unknown>) => {
          const task = tail.then(action);
          tail = task.catch(() => {});
          return task;
        },
      },
    },
  });
}
const unauthorized = () =>
  Response.json({ error: "authentication_required" }, { status: 401 });

test("two tabs with expired access perform one refresh and both succeed", async () => {
  installLocks();
  const first = await import(
    new URL("../src/features/auth/api.ts?tab=one", import.meta.url).href
  );
  const second = await import(
    new URL("../src/features/auth/api.ts?tab=two", import.meta.url).href
  );
  let active = false;
  let refreshes = 0;
  globalThis.fetch = async (input) => {
    if (String(input).endsWith("/auth/refresh")) {
      refreshes++;
      await new Promise((resolve) => setTimeout(resolve, 10));
      active = true;
      return new Response(null, { status: 204 });
    }
    return active ? Response.json({ id: "a" }) : unauthorized();
  };
  const result = await Promise.all([first.authApi.me(), second.authApi.me()]);
  assert.equal(refreshes, 1);
  assert.deepEqual(
    result.map((v: { id: string }) => v.id),
    ["a", "a"],
  );
});

test("legacy tab recovers losing refresh after winner updates cookies", async () => {
  Object.defineProperty(globalThis, "navigator", {
    configurable: true,
    value: {},
  });
  let calls = 0;
  globalThis.fetch = async () => {
    calls++;
    return calls < 4 ? unauthorized() : Response.json({ id: "a" });
  };
  assert.equal((await authApi.me()).id, "a");
});

test("requests carry displayed account and reject responses after account changes", async () => {
  setAccount("a");
  let release!: (response: Response) => void;
  globalThis.fetch = async (_input, init) => {
    assert.equal(
      (init?.headers as Record<string, string>)["X-Account-ID"],
      "a",
    );
    return new Promise((resolve) => {
      release = resolve;
    });
  };
  const operation = authenticatedRequest("/notifications");
  setAccount("b");
  release(Response.json([{ id: "a-private-notice" }]));
  await assert.rejects(
    operation,
    (error: { code: string }) => error.code === "account_changed",
  );
  assert.equal(accountSnapshot().id, "b");
});

test("terminal 401 invalidates current user and notifies session observers", async () => {
  installLocks();
  setAccount("a");
  let changes = 0;
  const stop = watchSession(() => {
    changes++;
  });
  globalThis.fetch = async () => unauthorized();
  await assert.rejects(authenticatedRequest("/notifications"));
  stop();
  assert.equal(accountSnapshot().id, null);
  assert.equal(changes, 1);
});

test("session invalidation clears account before observers run", () => {
  setAccount("a");
  const stop = watchSession(() => assert.equal(accountSnapshot().id, null));
  invalidateSession();
  stop();
});
