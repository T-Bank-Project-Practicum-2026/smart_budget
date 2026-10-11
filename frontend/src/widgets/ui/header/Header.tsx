import React from "react";
import styles from "./Header.module.scss";
import Button from "../../../shared/ui/Button";
import HeaderButton from "../../../shared/ui/HeaderButton";

export default function Header() {
  return (
    <header className={styles.container}>
      <HeaderButton endpoint="/butget">Бюжет</HeaderButton>
      <HeaderButton endpoint="/goals" notification={5}>
        Цели
      </HeaderButton>
      <HeaderButton endpoint="">Инвестиции</HeaderButton>
      <HeaderButton endpoint="">Уведомления</HeaderButton>
    </header>
  );
}
