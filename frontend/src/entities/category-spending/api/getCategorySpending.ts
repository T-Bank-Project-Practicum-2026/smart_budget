import { apiRequest } from "../../../shared/api";
import type { CategorySpending } from "../model/types";

export function getCategorySpendings(): Promise<CategorySpending[]> {
  return apiRequest<CategorySpending[]>("/category-spendings");
}
