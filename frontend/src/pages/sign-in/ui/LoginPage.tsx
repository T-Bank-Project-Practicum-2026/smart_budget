import React, { useState } from "react";
import Container from "../../../shared/ui/Container";
import styles from "./LoginPage.module.scss";
import Field from "../../../shared/ui/Field";
import Button from "../../../shared/ui/Button";
import LoginIcon from "../../../shared/ui/logo/key-round.svg";
import LockIcon from "../../../shared/ui/logo/lock-keyhole.svg";
import MailIcon from "../../../shared/ui/logo/mail.svg";
import { Link } from "react-router-dom";

export function AuthPage() {
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");

  return (
    <div className={styles.page}>
      <h1>Войти в аккаунт</h1>
      <form className={styles.container}>
        <div className={styles.title}>
          <h2>Умный бюджет</h2>
        </div>
        <div className={styles.input}>
          <Field
            icon={MailIcon}
            placeholder="Почта"
            type="email"
            value={email}
            onChange={setEmail}
          />
          <Field
            icon={LockIcon}
            placeholder="Пароль"
            type="password"
            value={password}
            onChange={setPassword}
          />
        </div>
        <Button image={LoginIcon}>Войти</Button>
        <Link to={"/register"} className={styles.register}>
          Нет аккаунта? Зарегистрироваться
        </Link>
      </form>
    </div>
  );
}
