import { POST } from "../client";

export interface LoginDto {
  email: string;
  password: string;
}

export interface LoginResponseDto {
  createdAt: string;
  email: string;
  id: string;
  timezone: string;
}

export function login(dto: LoginDto) {
  const options: RequestInit = {
    body: JSON.stringify(dto),
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
  };

  return POST<LoginResponseDto>("/v1/auth/login", options);
}
