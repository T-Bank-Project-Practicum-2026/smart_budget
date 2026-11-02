import React, { useState } from "react";
import Container from "../../../shared/ui/Container";
import styles from "./LoginPage.module.scss";
import Field from "../../../shared/ui/Field";
import Button from "../../../shared/ui/Button";
import LoginIcon from "../../../shared/ui/logo/key-round.svg";
import LockIcon from "../../../shared/ui/logo/lock-keyhole.svg";
import MailIcon from "../../../shared/ui/logo/mail.svg";
import { Link, useNavigate } from "react-router-dom";
import { useLogin } from "../../../features/auth/api/useLogin";

export function AuthPage() {
  const { mutate, isPending, error } = useLogin();
  const navigate = useNavigate();

  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");

  const handleSubmit = (e: React.SubmitEvent) => {
    e.preventDefault();

    mutate(
      { email, password },
      { onSuccess: () => navigate("/profile", { replace: true }) },
    );
  };

  return (
    <div className={styles.page}>
      <h1>Войти в аккаунт</h1>

      <form className={styles.container} onSubmit={handleSubmit}>
        {isPending && (
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
            <span>{error.message}</span>
          </div>
        )}

        <Button image={LoginIcon} type="submit" isDisabled={isPending}>
          Войти
        </Button>

        <Link to={"/register"} className={styles.register}>
          Нет аккаунта? Зарегистрироваться
        </Link>
      </form>
    </div>
  );
}
