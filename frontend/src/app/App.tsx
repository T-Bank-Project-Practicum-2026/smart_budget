import "./styles";
import AuthForm from "../features/auth/AuthForm";
import { useSession } from "../features/auth/useSession";
import PushPanel from "../features/push/PushPanel";
import Header from "../widgets/ui/header";
import BudgetPage from "../pages/BudgetPage";
import Footer from "../shared/ui/Footer";

function App() {
  const session = useSession();
  return (
    <>
      {session.user && <Header />}
      <main>
        <div className="session-panel">
          <h1>Умный бюджет</h1>
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
        </div>
        {!session.loading && session.user && <BudgetPage />}
      </main>
      <Footer />
    </>
  );
}

export default App;
