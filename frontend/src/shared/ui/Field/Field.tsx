import React from "react";
import styles from "./Field.module.scss";

interface FieldProps {
  placeholder: string;
  type: string;
  value: string;
  onChange: Function;
  icon: string;
}

export default function Field({
  placeholder,
  type,
  value,
  onChange,
  icon,
}: FieldProps) {
  return (
    <div className={styles.container}>
      {icon && <img src={icon} className={styles.icon} />}
      <input
        className={styles.input}
        placeholder={placeholder}
        type={type}
        autoComplete="off"
        value={value}
        onChange={(e) => {
          onChange(e.target.value);
          console.log(e.target.value);
        }}
      />
    </div>
  );
}
