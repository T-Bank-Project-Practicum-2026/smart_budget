import { useState } from "react";
import { ApiResult } from "../../../shared/api/client";
import { me, MeResponseDto } from "../../../shared/api/endpoints/me";

export default function useMe() {
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const handleMe = async (): Promise<ApiResult<MeResponseDto>> => {
    setIsLoading(true);
    setError(null);

    const result = await me();

    setIsLoading(false);

    if (!result.success) {
      setError(result.error || "Произошла ошибка при получении информации");
    }

    return result;
  };

  return {
    me: handleMe,
    isLoading,
    error,
  };
}
