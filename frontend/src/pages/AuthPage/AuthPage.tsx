import React from "react";
import Container from "../../shared/ui/Container";
import LoginForm from "../../features/auth/login";

export default function AuthPage() {
  return (
    <Container>
      <h1>Вход в аккаунт</h1>
      <LoginForm />
    </Container>
  );
}
