import React, { useEffect } from "react";
import styles from "./ProfilePage.module.scss";
import useMe from "../../../features/auth/api/useMe";
import { me } from "../../../shared/api/endpoints/me";

export function ProfilePage() {
  const { me, isLoading, error } = useMe();

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const result = await me();
        if (result.success) {
          console.log("Пользователь:", result.json);
        } else {
          console.error("Что то походу пошло не так :<(   :", result.error);
        }
      } catch (error) {
        console.error("Ошибка при запросе:", error);
      }
    };

    fetchUser();

    return () => {};
  }, []);

  return (
    <div className={styles.page}>
      <section className={styles.section}>
        <h1>Профиль</h1>
      </section>
      <section>
        <div className={styles.container}>
          <img
            className={styles.avatar}
            src="https://cdn-icons-png.magnific.com/256/25/25469.png?semt=ais_white_label"
          />
          <div className={styles.info_container}>
            <div className={styles.row}>
              <span>Почта</span>
              <span>example@mail.ru</span>
            </div>
            <div className={styles.row}>
              <span>Пароль</span>
              <span>Изменить</span>
            </div>
          </div>
        </div>

        {/* TODO: Кнопки (Выйти / История импортов выписок) */}
        {/* TODO: Логи событий вынести сюда */}
      </section>
    </div>
  );
}
