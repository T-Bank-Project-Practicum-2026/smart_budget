import React, { useState } from "react";
import styles from "./PageSelectMenu.module.scss";
import MenuLogo from "../../../shared/ui/logo/menu.svg";
import HeaderButton from "../../../shared/ui/HeaderButton";

export function PageSelectMenu() {
  const [isOpen, setIsOpen] = useState<boolean>(false);

  return (
    <div>
      {isOpen ? (
        <div className={styles.container}>
          <img src={MenuLogo} alt="Меню" onClick={() => setIsOpen(false)} />
          <div className={styles.selector}>
            <HeaderButton endpoint="/budget">Бюжет</HeaderButton>
            <HeaderButton endpoint="/goals" notification={5}>
              Цели
            </HeaderButton>
            {/* <HeaderButton endpoint="">Инвестиции</HeaderButton>
                        <HeaderButton endpoint="">Уведомления</HeaderButton> */}
          </div>
        </div>
      ) : (
        <img src={MenuLogo} alt="Меню" onClick={() => setIsOpen(true)} />
      )}
    </div>
  );
}
