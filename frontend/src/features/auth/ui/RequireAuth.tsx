import { Navigate, Outlet } from "react-router-dom";
import { useMe } from "../api/useMe";

export function RequireAuth() {
  const { data, error, isPending, isError } = useMe();

  if (error?.status === 401) return <Navigate to="/login" replace />;
  if (isPending) return null;
  if (isError && !data) {
    return <span>Не удалось проверить</span>;
  }

  return <Outlet />;
}
