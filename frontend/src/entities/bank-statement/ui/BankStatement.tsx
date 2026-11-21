import { Import } from "../../../shared/api/endpoints/imports";
import styles from "./BankStatement.module.scss";
import FileIcon from "../../../shared/ui/logo/file.svg";
import { formatDate } from "../../../shared/lib/format-date";

type BankStatementProps = {
  data: Import;
};

function BankStatement({ data }: BankStatementProps) {
  return (
    <div className={styles.container}>
      <div className={styles.title}>
        <img src={FileIcon} alt="" />
        <span>Информация о загруженном файле: {data.fileName}</span>
      </div>

      <div className={styles.row}>
        <span>Статус</span>
        <span>{data.status}</span>
      </div>

      <div className={styles.row}>
        <span>Успешно добавленных транзакций</span>
        <span>{data.importedRows}</span>
      </div>

      <div className={styles.row}>
        <span>Пропущенных транзаций</span>
        <span>{data.skippedRows}</span>
      </div>

      <div className={styles.row}>
        <span>Повторных транзакций</span>
        <span>{data.duplicateRows}</span>
      </div>

      <div className={styles.row}>
        <span>Конфликтующих транзакций</span>
        <span>{data.conflictRows}</span>
      </div>

      <div className={styles.row}>
        <span>Неуспешно добавленных транзакций</span>
        <span>{data.failedRows}</span>
      </div>

      <div className={styles.date}>
        <span>{formatDate(data.createdAt)}</span>
      </div>
    </div>
  );
}

export default BankStatement;
