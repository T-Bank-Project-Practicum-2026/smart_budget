import React from "react";
import styles from "./Badge.module.scss";

interface BadgeProps {
  children: React.ReactNode;
}

export default function Badge({ children }: BadgeProps) {
  return <div className={styles.container}>{children}</div>;
}
