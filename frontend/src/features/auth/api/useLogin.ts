import { useEffect, useState } from "react";
import {
    LoginDto,
    LoginResponseDto,
    login,
} from "../../../shared/api/endpoints/login";
import { ApiResult } from "../../../shared/api/client";

export default function useLogin() {
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);

    const handleLogin = async (
        dto: LoginDto,
    ): Promise<ApiResult<LoginResponseDto>> => {
        setIsLoading(true);
        setError(null);

        const result = await login(dto);

        setIsLoading(false);

        if (!result.success) {
            setError(result.error || "Произошла ошибка при входе");
        }

        return result;
    };

    return {
        login: handleLogin,
        isLoading,
        error,
    };
}
