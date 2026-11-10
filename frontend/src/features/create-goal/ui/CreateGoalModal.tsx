import { Modal } from "@/shared/ui/Modal/Modal";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./CreateGoalModal.module.scss";
import Button from "@/shared/ui/Button";
import SaveIcon from "../../../shared/ui/logo/save.svg";
import { useTbankImport } from "@/features/imports/api/useTbankImport";

export type CreateGoalModalProps = {
  isOpen: boolean;
  onClose: () => void;
};

export function CreateGoalModal({ isOpen, onClose }: CreateGoalModalProps) {
  const [title, setTitle] = useState<string>("");
  const [total, setTotal] = useState(10000);
  const [accumulated, setAccumulated] = useState(0);
  const [deadline, setDeadline] = useState<string>(
    new Date().toLocaleDateString(),
  );

  //   const { mutate, isPending, error } = useTbankImport();
  const navigate = useNavigate();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Создать цель">
      <form className={styles.content} onSubmit={handleSubmit}>
        <div className={styles.row}>
          <span>Название</span>
          <input
            placeholder="Название"
            type="text"
            onChange={(e) => {
              setTitle(e.target.value);
            }}
            value={title}
          />
        </div>

        <div className={styles.row}>
          <span>Необходимо накопить</span>
          <input
            placeholder="Сумма"
            type="number"
            onChange={(e) => {
              setAccumulated(Number(e.target.value));
            }}
            value={accumulated}
          />
        </div>

        <div className={styles.row}>
          <span>Уже накоплено</span>

          <input
            placeholder="Уже накоплено"
            type="number"
            onChange={(e) => {
              setTotal(Number(e.target.value));
            }}
            value={total}
          />
        </div>

        <div className={styles.row}>
          <span>Срок</span>

          <input
            placeholder="Срок"
            type="date"
            onChange={(e) => {
              setDeadline(e.target.value);
            }}
            value={deadline}
          />
        </div>

        <div className={styles.actions}>
          <Button type="submit" image={SaveIcon}>
            Сохранить
          </Button>
        </div>
      </form>
    </Modal>
  );
}
