import React, { useState } from "react";
import Button from "../../../shared/ui/Button";
import ImportBankStatementModal from "../../../features/importBankStatementModal/ui/ImportBankStatementModal";
import styles from "./ButtonActionButtons.module.scss";

import DownloadIcon from "../../../shared/ui/logo/download.svg";
import PlusIcon from "../../../shared/ui/logo/plus.svg";

type ModalType = "upload" | null;

export default function ButgetActionButtons() {
  const [openModal, setOpenModal] = useState<ModalType>(null);

  return (
    <div className={styles.container}>
      <Button onClick={() => setOpenModal("upload")} image={DownloadIcon}>
        Загрузить выписку
      </Button>
      <Button image={PlusIcon}>Добавить транзакцию</Button>

      <ImportBankStatementModal
        isOpen={openModal === "upload"}
        onClose={() => setOpenModal(null)}
      />
    </div>
  );
}
