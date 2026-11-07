import { useMutation, useQueryClient } from "@tanstack/react-query";
import { register } from "../../../shared/api/endpoints/register";
import { QueryKeys } from "../../../shared/api/queryKeys";
import { queryClient } from "../../../shared/api";

export function useRegister() {
  return useMutation({
    mutationFn: register,
    onSuccess: (user) => queryClient.setQueryData([QueryKeys.me], user),
  });
}
