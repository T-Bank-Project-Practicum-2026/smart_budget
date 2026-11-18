import { POST } from "../client";
import type { User } from "./me";

export type RegisterDto = {
  email: string;
  password: string;
  timezone: string;
};

export function register(dto: RegisterDto) {
  return POST<User>("/v1/auth/register", dto);
}
