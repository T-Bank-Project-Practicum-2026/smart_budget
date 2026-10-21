import { ReactNode, useEffect } from "react";
import { Portal } from "../Portal/Portal";
import styles from "./Modal.module.scss";

import CloseIcon from "./x.svg";

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  children: ReactNode;
  title?: string;
}

export const Modal = ({ isOpen, onClose, children, title }: ModalProps) => {
  useEffect(() => {
    if (!isOpen) return;

    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };

    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    document.addEventListener("keydown", onKeyDown);

    return () => {
      document.body.style.overflow = prevOverflow;
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <Portal>
      <div className={styles.overlay} onClick={onClose}>
        <div
          className={styles.content}
          onClick={(e) => e.stopPropagation()}
          role="dialog"
          aria-modal="true"
        >
          {title && (
            <header className={styles.header}>
              <h2>{title}</h2>
              <button
                onClick={onClose}
                aria-label="Закрыть"
                className={styles.close}
              >
                <img src={CloseIcon} />
              </button>
            </header>
          )}
          {children}
        </div>
      </div>
    </Portal>
  );
};
