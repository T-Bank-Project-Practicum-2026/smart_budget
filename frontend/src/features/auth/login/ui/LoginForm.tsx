import React from "react";
import Field from "../../../../shared/ui/Field";
import styles from "./LoginForm.module.scss";
import Button from "../../../../shared/ui/Button";

export default function LoginForm() {
  return (
    <div className={styles.container}>
      <div className={styles.input_container}>
        <h2>Умный бюджет</h2>
        <Field placeholder="Почта" type="email" />
        <Field placeholder="Пароль" type="password" />
        <Button>Войти</Button>
        <a>Нет аккаунта? Зарегестрироваться</a>
      </div>
    </div>
  );
}
