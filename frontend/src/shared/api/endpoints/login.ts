import { POST } from "../client";
import type { User } from "./me";

export type LoginDto = {
  email: string;
  password: string;
};

export function login(dto: LoginDto) {
  return POST<User>("/v1/auth/login", dto);
}
