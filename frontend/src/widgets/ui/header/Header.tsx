import React from "react";
import styles from "./Header.module.scss";
import Button from "../../../shared/ui/Button";
import HeaderButton from "../../../shared/ui/HeaderButton";

export default function Header() {
  return (
    <header className={styles.container}>
      <HeaderButton>Бюжет</HeaderButton>
      <HeaderButton notification={5}>Цели</HeaderButton>
      <HeaderButton>Инвестиции</HeaderButton>
      <HeaderButton>Уведомления</HeaderButton>
    </header>
  );
}
