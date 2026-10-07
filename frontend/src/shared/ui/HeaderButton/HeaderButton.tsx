import React from "react";
import styles from "./HeaderButton.module.scss";
import Badge from "../Badge";
import { Link } from "react-router-dom";

interface HeaderButtonProps {
  children: React.ReactNode;
  endpoint: string;
  notification?: number;
}

export default function HeaderButton({
  children,
  endpoint,
  notification,
}: HeaderButtonProps) {
  return (
    <Link to={endpoint} className={styles.container}>
      {children}
      {(notification ?? 0) > 0 ? <Badge>{notification}</Badge> : ""}
    </Link>
  );
}
