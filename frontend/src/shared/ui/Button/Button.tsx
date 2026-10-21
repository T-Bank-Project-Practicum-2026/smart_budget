import React from "react";
import styles from "./Button.module.scss";

interface ButtonProps {
  image?: string;
  children: React.ReactNode;
  isDisabled?: boolean;
  onClick?: React.MouseEventHandler<HTMLButtonElement>;
  type?: "button" | "submit" | "reset";
}

export default function Button({
  image,
  children,
  isDisabled = false,
  onClick,
  type = "button",
}: ButtonProps) {
  return (
    <button
      className={`${styles.button}`}
      type={type}
      disabled={isDisabled}
      onClick={onClick}
    >
      <div className={styles.content}>
        {image && <img src={image} />}
        {children}
      </div>
    </button>
  );
}
