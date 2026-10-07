import React from "react";
import styles from "./HeaderButton.module.scss";
import Badge from "../Badge";

interface HeaderButtonProps {
  children: React.ReactNode;
  notification?: number;
}

export default function HeaderButton({
  children,
  notification,
}: HeaderButtonProps) {
  return (
    <div className={styles.container}>
      <button className={styles.button}>{children}</button>
      {(notification ?? 0 > 0) ? <Badge>{notification}</Badge> : ""}
    </div>
  );
}
