import { POST } from "../client";

export type RegisterDto = {
    email: string;
    password: string;
    timezone: string;
};

export type RegisterResponseDto =
    | {
          createdAt: string;
          email: string;
          id: string;
          timezone: string;
      }
    | { error: string };

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
