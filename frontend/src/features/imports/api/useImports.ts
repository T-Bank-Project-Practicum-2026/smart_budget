import { useQuery } from "@tanstack/react-query";
import { imports } from "../../../shared/api/endpoints/imports";
import { QueryKeys } from "../../../shared/api/queryKeys";
import { useMe } from "../../auth/api/useMe";

export function useImports() {
  const accountId = useMe().data?.id ?? "";
  return useQuery({
    queryKey: [QueryKeys.imports, accountId],
    queryFn: () => imports(accountId),
    enabled: accountId !== "",
  });
}
