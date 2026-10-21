import React from "react";
import Container from "../../shared/ui/Container";
import Goal from "../../shared/ui/Goal";

export default function GoalsPage() {
  return (
    <Container>
      <h1>Мои цели</h1>
      <Goal />
      <Goal />
      <Goal />
    </Container>
  );
}
