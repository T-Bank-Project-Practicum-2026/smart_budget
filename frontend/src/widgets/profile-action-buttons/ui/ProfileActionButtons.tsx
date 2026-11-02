import React, { useState } from "react";
import Button from "../../../shared/ui/Button";
import ImportBankStatementModal from "../../../features/importBankStatementModal/ui/ImportBankStatementModal";
import styles from "./ProfileActionButtons.module.scss";

import DownloadIcon from "../../../shared/ui/logo/download.svg";
import PlusIcon from "../../../shared/ui/logo/plus.svg";
import LogoutIcon from "../../../shared/ui/logo/log-out.svg";
import { useLogout } from "../../../features/auth/api/useLogout";
import { useNavigate } from "react-router-dom";

type ModalType = "upload" | null;

export default function ProfileActionButtons() {
  const [openModal, setOpenModal] = useState<ModalType>(null);
  const navigate = useNavigate();

  const { mutate, isPending, error } = useLogout();

  const logoutOnClick = () => {
    mutate(undefined, {
      onSuccess: () => navigate("/", { replace: true }),
    });
  };

  return (
    <div className={styles.container}>
      <Button onClick={() => setOpenModal("upload")} image={DownloadIcon}>
        Загрузить выписку
      </Button>
      <Button image={PlusIcon}>Добавить транзакцию</Button>
      <Button image={LogoutIcon} onClick={() => logoutOnClick()}>
        Выйти
      </Button>

      <ImportBankStatementModal
        isOpen={openModal === "upload"}
        onClose={() => setOpenModal(null)}
      />
    </div>
  );
}
