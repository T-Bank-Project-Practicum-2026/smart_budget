import React from "react";
import Container from "../../../shared/ui/Container";
import Goal from "../../../shared/ui/Goal";
import styles from "./GoalsPage.module.scss";

export function GoalsPage() {
  return (
    <div className={styles.page}>
      <section>
        <h1>Мои цели</h1>
      </section>
      <Goal />
      <Goal />
      <Goal />
    </div>
  );
}
