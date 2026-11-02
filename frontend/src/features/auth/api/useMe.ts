import { useQuery } from "@tanstack/react-query";
import { me } from "../../../shared/api/endpoints/me";
import { QueryKeys } from "../../../shared/api/queryKeys";

export function useMe() {
  return useQuery({ queryKey: [QueryKeys.me], queryFn: me });
}
