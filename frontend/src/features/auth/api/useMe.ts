import { useEffect, useState } from "react";
import { me, MeResponseDto } from "../../../shared/api/endpoints/me";

export type UseMe = {
    data: MeResponseDto | null;
    isLoading: boolean;
    error: string | null;
};

export default function useMe() {
    const [data, setData] = useState<MeResponseDto | null>(null);
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchData = async () => {
            setIsLoading(true);
            setError(null);

            const result = await me();

            if (result.success) {
                setData(result.json);
            } else {
                setError(result.error);
            }
            setIsLoading(false);
        };

        fetchData();
    }, []);

    return { data, isLoading, error } as UseMe;
}
