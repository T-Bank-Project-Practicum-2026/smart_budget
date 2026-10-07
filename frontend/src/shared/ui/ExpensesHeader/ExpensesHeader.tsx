import React from "react";
import styles from "./ExpensesHeader.module.scss";
import { formatCurrency } from "../../lib/format-currency";
import CategoryBadge from "../CategoryBadge";
import image from "./chevron-down.svg";

export default function ExpensesHeader() {
  return (
    <div className={styles.container}>
      <div className={styles.title}>
        <div>
          <img src={image} />
        </div>
        <h2>Сегодня</h2>
        <div className={styles.categories}>
          <CategoryBadge
            className={styles.category}
            style={{ zIndex: 3 }}
            color="#ff3399"
          />
          <CategoryBadge
            className={styles.category}
            style={{ zIndex: 2 }}
            color="#00cccc"
          />
          <CategoryBadge
            className={styles.category}
            style={{ zIndex: 1 }}
            color="#d1d1d1"
          />
        </div>
      </div>
      <a>- {formatCurrency(14000)}</a>
    </div>
  );
}
