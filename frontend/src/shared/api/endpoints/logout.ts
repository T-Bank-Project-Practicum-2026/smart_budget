import { POST } from "../client";

export function logout(accountId: string) {
  return POST<void>("/v1/auth/logout", undefined, accountId);
}
