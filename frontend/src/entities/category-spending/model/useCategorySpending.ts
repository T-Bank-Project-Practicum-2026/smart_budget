import { useEffect, useState } from "react";
import { getCategorySpendings } from "../api/getCategorySpending";
import type { CategorySpending } from "./types";

export function useCategorySpendings() {
  const [categorySpendings, setCategorySpendings] = useState<
    CategorySpending[]
  >([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    getCategorySpendings()
      .then((data) => {
        if (!cancelled) setCategorySpendings(data.json);
      })
      .catch((e: Error) => {
        if (!cancelled) setError(e.message);
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  return { categorySpendings, isLoading, error };
}
