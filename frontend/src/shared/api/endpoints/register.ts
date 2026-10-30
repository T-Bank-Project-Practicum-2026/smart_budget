import { POST } from "../client";

export interface RegisterDto {
  email: string;
  password: string;
  timezone: string;
}

export interface RegisterResponseDto {
  createdAt: string;
  email: string;
  id: string;
  timezone: string;
}

export function register(dto: RegisterDto) {
  const options: RequestInit = {
    body: JSON.stringify(dto),
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
  };

  return POST<RegisterResponseDto>("/v1/auth/register", options);
}
