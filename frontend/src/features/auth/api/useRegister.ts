import { useState } from "react";
import { ApiResult } from "../../../shared/api/client";
import {
  register,
  RegisterDto,
  RegisterResponseDto,
} from "../../../shared/api/endpoints/register";

export default function useRegister() {
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const handleRegister = async (
    dto: RegisterDto,
  ): Promise<ApiResult<RegisterResponseDto>> => {
    setIsLoading(true);
    setError(null);

    const result = await register(dto);

    setIsLoading(false);

    if (!result.success) {
      setError(result.error || "Произошла ошибка при регистрации");
    }

    return result;
  };

  return {
    register: handleRegister,
    isLoading,
    error,
  };
}
