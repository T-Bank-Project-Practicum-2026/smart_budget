import { useQuery } from "@tanstack/react-query";
import { imports } from "../../../shared/api/endpoints/imports";
import { QueryKeys } from "../../../shared/api/queryKeys";

export function useImports() {
  return useQuery({ queryKey: [QueryKeys.imports], queryFn: imports });
}
