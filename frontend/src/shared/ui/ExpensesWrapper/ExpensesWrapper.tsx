import React from "react";
import styles from "./ExpensesWrapper.module.scss";
import ExpensesHeader from "../ExpensesHeader";
import ExpensesBody from "../ExpensesBody";

export default function ExpensesWrapper() {
  return (
    <div className={styles.container}>
      <ExpensesHeader />
      <div className={styles.content}>
        <ExpensesBody />
        <ExpensesBody />
        <ExpensesBody />
        <ExpensesBody />
      </div>
    </div>
  );
}
