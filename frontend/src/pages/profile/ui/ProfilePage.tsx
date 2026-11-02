import React from "react";
import { Navigate } from "react-router-dom";
import styles from "./ProfilePage.module.scss";
import UserIcon from "../../../shared/ui/logo/user-round.svg";
import { useMe } from "../../../features/auth/api/useMe";
import { useImports } from "../../../features/imports/api/useImports";
import { formatDate } from "../../../shared/lib/format-date";
import BankStatement from "../../../entities/bank-statement/ui/BankStatement";
import BankStatementSkeleton from "../../../entities/bank-statement/ui/BankStatementSkeleton";
import ProfileActionButtons from "../../../widgets/profile-action-buttons/ui/ProfileActionButtons";

export function ProfilePage() {
  const me = useMe();
  const imports = useImports();

  if (me.error?.status === 401) return <Navigate to="/login" replace />;

  return (
    <div className={styles.page}>
      <section className={styles.section}>
        <h1>Профиль</h1>
      </section>

      {me.isPending && <div className={styles.skeleton_container} />}
      {me.isError && (
        <span>Ошибка при загрузке профиля: {me.error.message}</span>
      )}
      {me.data && (
        <section>
          <div className={styles.container}>
            <img className={styles.avatar} src={UserIcon} alt="" />
            <div className={styles.info_container}>
              <div className={styles.row}>
                <span>Почта</span>
                <span>{me.data.email}</span>
              </div>
              <div className={styles.row}>
                <span>ID</span>
                <span>{me.data.id}</span>
              </div>
              <div className={styles.row}>
                <span>Дата регистрации</span>
                <span>{formatDate(me.data.createdAt)}</span>
              </div>
              <div className={styles.row}>
                <span>Временная зона</span>
                <span>{me.data.timezone}</span>
              </div>
            </div>
          </div>
        </section>
      )}

      <ProfileActionButtons />

      <section className={styles.section}>
        <h2>Загруженные выписки</h2>
      </section>

      {imports.isPending &&
        [1, 2, 3].map((i) => <BankStatementSkeleton key={i} />)}
      {imports.isError && <div>{imports.error.message}</div>}
      {imports.data?.length === 0 && <span>Выписок пока нет</span>}
      {imports.data?.map((i) => (
        <BankStatement key={i.id} data={i} />
      ))}

      <section className={styles.section}></section>
    </div>
  );
}
