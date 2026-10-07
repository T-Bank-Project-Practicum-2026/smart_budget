import { useState, type FormEvent } from "react";
import { errorMessage } from "../../api";
import { authApi, type User } from "./api";

export default function AuthForm({
  onSuccess,
}: {
  onSuccess: (user: User) => void;
}) {
  const [register, setRegister] = useState(false);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const data = new FormData(form);
    const credentials = {
      email: String(data.get("email")),
      password: String(data.get("password")),
    };
    setBusy(true);
    setMessage("");
    try {
      const user = register
        ? await authApi.register({
            ...credentials,
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
          })
        : await authApi.login(credentials);
      form.reset();
      onSuccess(user);
    } catch (error) {
      setMessage(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }

  return (
    <section>
      <h2>{register ? "Регистрация" : "Вход"}</h2>
      <form onSubmit={(event) => void submit(event)}>
        <label>
          Email
          <input
            name="email"
            type="email"
            autoComplete="username"
            maxLength={254}
            required
            disabled={busy}
          />
        </label>
        <label>
          Пароль
          <input
            name="password"
            type="password"
            autoComplete={register ? "new-password" : "current-password"}
            minLength={8}
            required
            disabled={busy}
          />
        </label>
        {register && (
          <small>
            Не менее 8 символов. Часовой пояс определяется браузером.
          </small>
        )}
        <button type="submit" disabled={busy}>
          {busy ? "Подождите…" : register ? "Зарегистрироваться" : "Войти"}
        </button>
        <button
          type="button"
          className="secondary"
          disabled={busy}
          onClick={() => {
            setRegister(!register);
            setMessage("");
          }}
        >
          {register ? "Уже есть аккаунт" : "Создать аккаунт"}
        </button>
      </form>
      <p role="status">{message}</p>
    </section>
  );
}
