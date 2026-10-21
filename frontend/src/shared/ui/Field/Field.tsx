import React from "react";
import styles from "./Field.module.scss";

interface FieldProps {
  placeholder: string;
  type: string;
}

export default function Field({ placeholder, type }: FieldProps) {
  return (
    <input
      className={styles.input}
      placeholder={placeholder}
      type={type}
      autoComplete="off"
    />
  );
}
