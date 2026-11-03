import { useMutation, useQueryClient } from "@tanstack/react-query";
import { QueryKeys } from "../../../shared/api/queryKeys";
import { logout } from "../../../shared/api/endpoints/logout";
import { queryClient } from "../../../shared/api";

export function useLogout() {
  return useMutation({
    mutationKey: [QueryKeys.me],
    mutationFn: logout,
    onSuccess: () => queryClient.clear(),
  });
}
