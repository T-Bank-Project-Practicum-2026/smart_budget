import { Navigate, Outlet } from "react-router-dom";
import { useMe } from "../api/useMe";
import styles from "./RequireAuth.module.scss";

export function RequireAuth() {
  const { data, error, isPending, isError } = useMe();

  if (error?.status === 401) return <Navigate to="/login" replace />;
  if (isPending)
    return (
      <div className={styles.page}>
        <div className={styles.loader}></div>
      </div>
    );
  if (isError && !data) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}
