import { POST } from "../client";

export function logout() {
  return POST<void>("/v1/auth/logout");
}
