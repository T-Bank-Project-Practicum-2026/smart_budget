import { GET } from "../../../shared/api";
import { ApiResult } from "../../../shared/api/client";
import type { CategorySpending } from "../model/types";

export function getCategorySpendings(): Promise<ApiResult<CategorySpending[]>> {
  return GET<CategorySpending[]>("/category-spendings");
}
