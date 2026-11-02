import { useMutation, useQueryClient } from "@tanstack/react-query";
import { login } from "../../../shared/api/endpoints/login";
import { QueryKeys } from "../../../shared/api/queryKeys";

export function useLogin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationKey: [QueryKeys.me],
    mutationFn: login,
    onSuccess: (user) => queryClient.setQueryData([QueryKeys.me], user),
  });
}
