import React from "react";
import styles from "./ExpensesBody.module.scss";
import { formatCurrency } from "../../lib/format-currency";
import { CategoryBadge } from "../../../entities/category";

export default function ExpensesBody() {
  return (
    <div className={styles.container}>
      <div className={styles.info}>
        <div className={styles.logo}>
          <img src="https://krasnoeibeloe.servicecdn.ru/upload/resize_cache/iblock/14c/hxyzlkso2vi7qz709wvyrogckwpmhhth/1200_800_1/Red_icon.jpg" />
        </div>
        <div className={styles.content}>
          <span>Красное и белое</span>
          <CategoryBadge title="Супермаркеты" color="#33ff00" />
        </div>
      </div>

      <a>- {formatCurrency(1200)}</a>
    </div>
  );
}
