import React, { useState } from "react";
import { Modal } from "../../../shared/ui/Modal/Modal";
import Button from "../../../shared/ui/Button";
import styles from "./ImportBankStatementModal.module.scss";
import SaveIcon from "../../../shared/ui/logo/save.svg";
import HistoryIcon from "../../../shared/ui/logo/history.svg";
import TrashIcon from "../../../shared/ui/logo/trash.svg";
import { useNavigate } from "react-router-dom";

interface ImportBankStatementModalProps {
  isOpen: boolean;
  onClose: () => void;
}

function getFileSize(size: number) {
  if (size < 1024) {
    return size + "bytes";
  } else if (size > 1024 && size < 1048576) {
    return (size / 1024).toFixed(1) + "KB";
  } else if (size > 1048576) {
    return (size / 1048576).toFixed(1) + "MB";
  }
}

export default function ImportBankStatementModal({
  isOpen,
  onClose,
}: ImportBankStatementModalProps) {
  const [file, setFile] = useState<File | null>(null);
  const navigate = useNavigate();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onClose();
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    setFile(file ? file : null);
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Загрузить выписку">
      <form className={styles.content} onSubmit={handleSubmit}>
        <label className={styles.dropzone} htmlFor="statement-file">
          {file?.name ?? "Файл не выбран"}
        </label>
        <input
          id="statement-file"
          type="file"
          accept=".csv,.pdf,.xlsx"
          onChange={handleChange}
          className={styles.fileInput}
        />
        {file && <span>Размер: {getFileSize(file.size)}</span>}

        <div className={styles.actions}>
          <Button
            onClick={() => navigate("/import-history")}
            image={HistoryIcon}
          />
          <Button
            isDisabled={!file}
            onClick={() => setFile(null)}
            image={TrashIcon}
          />
          <Button type="submit" isDisabled={!file} image={SaveIcon}>
            Сохранить
          </Button>
        </div>
      </form>
    </Modal>
  );
}
