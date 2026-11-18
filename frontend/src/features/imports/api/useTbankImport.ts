import { useMutation } from "@tanstack/react-query";
import { importTbank } from "../../../shared/api/endpoints/tbank";
import { QueryKeys } from "../../../shared/api/queryKeys";
import { ApiError, queryClient } from "../../../shared/api";
import { useMe } from "../../auth/api/useMe";

export function useTbankImport() {
  const account = useMe().data;

  return useMutation({
    mutationKey: [QueryKeys.importTBank],
    mutationFn: async (file: File) => {
      if (!account) throw new ApiError(401, "authentication_required");
      await importTbank(file, account.id);
      return account.id;
    },
    onSuccess: (accountId) =>
      queryClient.invalidateQueries({
        queryKey: [QueryKeys.imports, accountId],
      }),
  });
}
