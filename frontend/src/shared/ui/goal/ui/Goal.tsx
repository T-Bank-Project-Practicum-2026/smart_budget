import React from "react";
import { formatCurrency } from "../../../lib/format-currency";
import styles from "./Goal.module.scss";
import CalendarClock from "@shared/ui/logo/calendar-clock.svg";
import BanknoteArrowDown from "@shared/ui/logo/banknote-arrow-down.svg";
import Wallet from "@shared/ui/logo/wallet.svg";
import Trash from "@shared/ui/logo/trash.svg";
import Edit from "@shared/ui/logo/square-pen.svg";
import Button from "../../Button";
import { CategoryBadge } from "@/entities/category";
import Percent from "@shared/ui/logo/percent.svg";
import HandCoins from "@shared/ui/logo/hand-coins.svg";
import Plus from "@shared/ui/logo/plus.svg";

export type GoalProps = {
  image: string;
  title: string;
  deadline: number;
  accumulated: number;
  total: number;
  monthly: number;
  isExample?: boolean;
};

export default function Goal({
  image,
  title,
  deadline,
  accumulated,
  monthly,
  total,
  isExample = false,
}: GoalProps) {
  return (
    <div className={styles.container}>
      <div className={styles.image}>
        <img src={image} />
      </div>
      <div className={styles.title}>
        <h3>{title}</h3>
        <CategoryBadge image={Percent} color="#FF0000" title="12.43/100.00" />
      </div>
      <div className={styles.info_container}>
        <div className={styles.row}>
          <div className={styles.row_title}>
            <img src={CalendarClock} />
            <span>Срок:</span>
          </div>
          <span>
            {new Date(deadline * 1000).toLocaleDateString("ru-RU", {
              day: "numeric",
              month: "numeric",
              year: "numeric",
            })}
          </span>
        </div>
        <div className={styles.row}>
          <div className={styles.row_title}>
            <img src={BanknoteArrowDown} />
            <span>Накоплено:</span>
          </div>
          <span>{formatCurrency(accumulated)}</span>
        </div>
        <div className={styles.row}>
          <div className={styles.row_title}>
            <img src={Wallet} />
            <span>Цель:</span>
          </div>
          <span>{formatCurrency(total)}</span>
        </div>
        <div className={styles.row}>
          <div className={styles.row_title}>
            <img src={HandCoins} />
            <span>Рек. взнос:</span>
          </div>
          <span>{formatCurrency(monthly)} / месяц</span>
        </div>
      </div>

      <div className={styles.progressbar}>
        <div style={{ width: "3%" }} />
      </div>

      {isExample === true && (
        <Button image={Plus} style={styles.example_button}>
          Добавить цель
        </Button>
      )}

      {isExample === false && (
        <div className={styles.buttons}>
          <Button image={BanknoteArrowDown} />
          <Button image={Edit} />
          <Button image={Trash} />
        </div>
      )}
    </div>
  );
}
