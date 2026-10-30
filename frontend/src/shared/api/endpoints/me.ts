import { GET } from "../client";

export interface MeResponseDto {
  createdAt: string;
  email: string;
  id: string;
  timezone: string;
}

export function me() {
  const options: RequestInit = {
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
  };

  return GET<MeResponseDto>("/v1/auth/me", options);
}
