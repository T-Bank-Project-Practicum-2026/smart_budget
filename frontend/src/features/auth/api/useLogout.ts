import { useMutation, useQueryClient } from "@tanstack/react-query";
import { QueryKeys } from "../../../shared/api/queryKeys";
import { logout } from "../../../shared/api/endpoints/logout";

export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationKey: [QueryKeys.me],
    mutationFn: logout,
    onSuccess: () => queryClient.removeQueries({ queryKey: [QueryKeys.me] }),
  });
}
