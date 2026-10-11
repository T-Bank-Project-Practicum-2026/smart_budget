import React from "react";
import styles from "./CategoryBadge.module.scss";
import ShoppingCart from "./shopping-cart.svg";
import { formatCurrency } from "../../lib/format-currency";

interface CategoryBadgeProps {
  color: string;
  title?: string;
  value?: number;
  className?: string;
  style?: React.CSSProperties;
}

export default function CategoryBadge({
  color,
  title,
  value,
  className,
  style,
}: CategoryBadgeProps) {
  return (
    <div
      style={{
        backgroundColor: `${color}50`,
        border: `1px solid ${color}75`,
        ...style,
      }}
      className={`${styles.container} ${className ?? ""}`}
    >
      <div style={{ backgroundColor: color }} className={styles.icon_container}>
        <img className={styles.icon} src={ShoppingCart} />
      </div>
      {title ? <a className={styles.title}>{title}</a> : ""}
      {value ? <a className={styles.value}>{formatCurrency(value)}</a> : ""}
    </div>
  );
}
