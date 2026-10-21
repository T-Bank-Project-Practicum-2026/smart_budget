import { apiRequest } from "../../../shared/api";
import type { Category } from "../model/types";

export function getCategories(): Promise<Category[]> {
  return apiRequest<Category[]>("/categories");
}
