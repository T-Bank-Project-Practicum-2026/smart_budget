import AuthForm from "./features/auth/AuthForm";
import { useSession } from "./features/auth/useSession";
import PushPanel from "./features/push/PushPanel";

export default function App() {
  const session = useSession();
  return (
    <main>
      <h1>Умный бюджет</h1>
      <p className="description">Учёт расходов и накоплений</p>
      {session.loading ? (
        <p>Загрузка…</p>
      ) : session.user ? (
        <>
          <section className="account">
            <p>{session.user.email}</p>
            <button
              disabled={session.busy}
              onClick={() => void session.logout()}
            >
              Выйти
            </button>
          </section>
          <PushPanel key={session.user.id} disabled={session.busy} />
        </>
      ) : (
        <AuthForm onSuccess={session.acceptUser} />
      )}
      <p role="status">{session.message}</p>
    </main>
  );
}
