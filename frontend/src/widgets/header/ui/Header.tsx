import React from "react";
import styles from "./Header.module.scss";
import Button from "../../../shared/ui/Button";
import HeaderButton from "../../../shared/ui/HeaderButton";
import { PageSelectMenu } from "./PageSelectMenu";

export function Header() {
  return (
    <header className={styles.background}>
      <div className={styles.container}>
        <div className={styles.logo}>
          <HeaderButton endpoint="/landing">Умный бюджет</HeaderButton>
        </div>
        <div className={styles.buttons}>
          <HeaderButton endpoint="/budget">Бюджет</HeaderButton>
          <HeaderButton endpoint="/goals" notification={5}>
            Цели
          </HeaderButton>
          {/* <HeaderButton endpoint="">Инвестиции</HeaderButton>
                    <HeaderButton endpoint="">Уведомления</HeaderButton> */}
        </div>
        <div className={styles.profile}>
          <HeaderButton endpoint="/profile">Профиль</HeaderButton>
        </div>
        <div className={styles.menu}>
          <PageSelectMenu />
        </div>
      </div>
    </header>
  );
}
