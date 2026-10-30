import { GET } from "../../../shared/api";
import { ApiResult } from "../../../shared/api/client";
import type { Category } from "../model/types";

export function getCategories(): Promise<ApiResult<Category[]>> {
  return GET<Category[]>("/categories");
}
