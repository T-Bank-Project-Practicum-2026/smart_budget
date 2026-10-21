import React from "react";
import { formatCurrency } from "../../lib/format-currency";
import styles from "./Goal.module.scss";

export default function Goal() {
  return (
    <div className={styles.container}>
      <h3 className={styles.title}>Квартира</h3>
      <div className={styles.value_container}>
        <a>{formatCurrency(1000000)}</a>
        <a>{formatCurrency(15000000)}</a>
      </div>
      <progress className={styles.progressbar} value={0.03}></progress>
    </div>
  );
}
