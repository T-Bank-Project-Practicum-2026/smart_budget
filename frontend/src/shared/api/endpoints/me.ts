import { GET } from "../client";

export type MeResponseDto = {
    createdAt: string;
    email: string;
    id: string;
    timezone: string;
};

export function me() {
    const options: RequestInit = {
        headers: {
            Accept: "application/json",
        },
    };

    return GET<MeResponseDto>("/v1/auth/me", options);
}
