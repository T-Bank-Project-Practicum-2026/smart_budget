import { GET } from "../client";

export type User = {
  createdAt: string;
  email: string;
  id: string;
  timezone: string;
};

export function me() {
  return GET<User>("/v1/auth/me");
}
