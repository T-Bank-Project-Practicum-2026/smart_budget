import React, { useState } from "react";
import Container from "../../../shared/ui/Container";
import styles from "./LoginPage.module.scss";
import Field from "../../../shared/ui/Field";
import Button from "../../../shared/ui/Button";
import LoginIcon from "../../../shared/ui/logo/key-round.svg";
import LockIcon from "../../../shared/ui/logo/lock-keyhole.svg";
import MailIcon from "../../../shared/ui/logo/mail.svg";
import { Link, useNavigate } from "react-router-dom";
import { useRegister } from "../../../features/auth/api/useRegister";

export function RegisterPage() {
  const navigate = useNavigate();

  const { mutate, isPending, error } = useRegister();

  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [timezone, setTimezone] = useState<string>("UTC");

  const handleSubmit = (e: React.SubmitEvent) => {
    e.preventDefault();

    mutate(
      { email, password, timezone },
      { onSuccess: () => navigate("/profile", { replace: true }) },
    );
  };

  return (
    <div className={styles.page}>
      <h1>Зарегистрировать аккаунт</h1>
      <form className={styles.container} onSubmit={handleSubmit}>
        {isPending && (
          <div className={styles.loading}>
            {/* Обновить на крутящуюся штучку */}
            <span>Происходит регестрация...</span>
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
          Зарегистрировать
        </Button>
        <Link to={"/login"} className={styles.register}>
          Есть аккаунт? Войти
        </Link>
      </form>
    </div>
  );
}
