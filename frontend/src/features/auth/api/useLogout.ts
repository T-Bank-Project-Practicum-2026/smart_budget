import { useMutation } from "@tanstack/react-query";
import { QueryKeys } from "../../../shared/api/queryKeys";
import { logout } from "../../../shared/api/endpoints/logout";
import { ApiError, queryClient } from "../../../shared/api";

import { useMe } from "./useMe";

export function useLogout() {
  const { data: account } = useMe();

  return useMutation({
    mutationKey: [QueryKeys.me],
    mutationFn: () => {
      if (!account) throw new ApiError(401, "authentication_required");
      return logout(account.id);
    },
    onSuccess: () => queryClient.clear(),
  });
}
