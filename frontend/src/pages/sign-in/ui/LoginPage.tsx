import React, { useState } from "react";
import Container from "../../../shared/ui/Container";
import styles from "./LoginPage.module.scss";
import Field from "../../../shared/ui/Field";
import Button from "../../../shared/ui/Button";
import LoginIcon from "../../../shared/ui/logo/key-round.svg";
import LockIcon from "../../../shared/ui/logo/lock-keyhole.svg";
import MailIcon from "../../../shared/ui/logo/mail.svg";
import { Link } from "react-router-dom";
import useLogin from "../../../features/auth/api/useLogin";

export function AuthPage() {
  const { login, isLoading, error } = useLogin();

  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");

  const handleSubmit = async (e: React.SubmitEvent) => {
    e.preventDefault();

    const result = await login({ email, password });

    if (result.success) {
      console.log("Вошел:", result.json);
    } else {
      console.error("Не вошел:", result.error);
    }
  };

  return (
    <div className={styles.page}>
      <h1>Войти в аккаунт</h1>

      <form className={styles.container} onSubmit={handleSubmit}>
        {isLoading && (
          <div className={styles.loading}>
            {/* Обновить на крутящуюся штучку */}
            <span>Происходит вход...</span>
          </div>
        )}

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

        {error && (
          <div className={styles.error}>
            <span>Произошла ошибка</span>
            <br />
            <span>{error}</span>
          </div>
        )}

        <Button image={LoginIcon} type="submit" isDisabled={isLoading}>
          Войти
        </Button>

        <Link to={"/register"} className={styles.register}>
          Нет аккаунта? Зарегистрироваться
        </Link>
      </form>
    </div>
  );
}
