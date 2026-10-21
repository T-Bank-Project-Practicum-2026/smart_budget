import React, { useState } from "react";
import { Modal } from "../../../shared/ui/Modal/Modal";
import Button from "../../../shared/ui/Button";
import styles from "./ImportBankStatementModal.module.scss";

interface ImportBankStatementModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function ImportBankStatementModal({
  isOpen,
  onClose,
}: ImportBankStatementModalProps) {
  const [fileName, setFileName] = useState<string | null>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onClose();
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    setFileName(file ? file.name : null);
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Загрузить выписку">
      <form className={styles.content} onSubmit={handleSubmit}>
        <label className={styles.dropzone} htmlFor="statement-file">
          {fileName ?? "Файл не выбран"}
        </label>
        <input
          id="statement-file"
          type="file"
          accept=".csv,.pdf,.xlsx"
          onChange={handleChange}
          className={styles.fileInput}
        />

        <div className={styles.actions}>
          <Button type="submit" isDisabled={!fileName}>
            Сохранить
          </Button>
        </div>
      </form>
    </Modal>
  );
}
