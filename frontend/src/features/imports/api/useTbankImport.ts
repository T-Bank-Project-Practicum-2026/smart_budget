import { useMutation, useQueryClient } from "@tanstack/react-query";
import { importTbank } from "../../../shared/api/endpoints/tbank";
import { QueryKeys } from "../../../shared/api/queryKeys";
import { Import } from "../../../shared/api/endpoints/imports";
import { queryClient } from "../../../shared/api";

export function useTbankImport() {
  return useMutation({
    mutationKey: [QueryKeys.importTBank],
    mutationFn: importTbank,
    onSuccess: (newItem) =>
      queryClient.setQueryData([QueryKeys.imports], (prev: Import[]) => [
        ...prev,
        newItem,
      ]),
  });
}
