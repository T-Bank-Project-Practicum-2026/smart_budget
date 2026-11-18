import React from "react";
import styles from "./Button.module.scss";

interface ButtonProps {
  image?: string;
  children?: React.ReactNode;
  isDisabled?: boolean;
  onClick?: React.MouseEventHandler<HTMLButtonElement>;
  onSubmit?: React.SubmitEventHandler;
  type?: "button" | "submit" | "reset";
}

export default function Button({
  image,
  children,
  isDisabled = false,
  onClick,
  onSubmit,
  type = "button",
}: ButtonProps) {
  return (
    <button
      className={`${styles.button}`}
      type={type}
      disabled={isDisabled}
      onClick={onClick}
      onSubmit={onSubmit}
    >
      <div className={styles.content}>
        {image && <img src={image} />}
        {children ? children : ""}
      </div>
    </button>
  );
}
