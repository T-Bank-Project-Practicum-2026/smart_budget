import React from "react";
import styles from "./ProfilePage.module.scss";
import useMe from "../../../features/auth/api/useMe";
import UserIcon from "../../../shared/ui/logo/user-round.svg";

export function ProfilePage() {
    const { data, isLoading, error } = useMe();

    return (
        <div className={styles.page}>
            <section className={styles.section}>
                <h1>Профиль</h1>
            </section>

            {isLoading && <div className={styles.skeleton_container} />}
            {data && (
                <section>
                    <div className={styles.container}>
                        <div className={styles.avatar_container}>
                            <img
                                className={styles.avatar}
                                src={UserIcon}
                                alt=""
                            />
                        </div>
                        <div className={styles.info_container}>
                            <div className={styles.row}>
                                <span>Почта</span>
                                <span>{data.email}</span>
                            </div>
                            <div className={styles.row}>
                                <span>ID</span>
                                <span>{data.id}</span>
                            </div>
                            <div className={styles.row}>
                                <span>Дата регистрации</span>
                                <span>{data.createdAt}</span>
                            </div>
                            <div className={styles.row}>
                                <span>Временная зона</span>
                                <span>{data.timezone}</span>
                            </div>
                        </div>
                    </div>
                </section>
            )}

            {error && <span>Ошибка при загрузке страницы: {error}</span>}
        </div>
    );
}
