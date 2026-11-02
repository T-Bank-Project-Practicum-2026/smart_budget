import React from "react";
import styles from "./Footer.module.scss";
import HeaderButton from "../HeaderButton";

export default function Footer() {
  return (
    <footer className={styles.container}>
      <HeaderButton endpoint="/login">Вход</HeaderButton>
      <HeaderButton endpoint="/register">Регистрация</HeaderButton>
    </footer>
  );
}
