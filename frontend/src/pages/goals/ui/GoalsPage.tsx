import React, { useState } from "react";
import Container from "../../../shared/ui/Container";
import Goal from "../../../shared/ui/goal";
import styles from "./GoalsPage.module.scss";
import Button from "@/shared/ui/Button";
import Plus from "@shared/ui/logo/plus.svg";

import GoalCar from "@shared/ui/logo/goal_car.png";
import GoalFlat from "@shared/ui/logo/goal_flat.png";
import { CreateGoalModal } from "@/features/create-goal";

export function GoalsPage() {
  const [isOpenModal, setIsOpenModal] = useState<boolean>(false);

  return (
    <div className={styles.page}>
      <section className={styles.title}>
        <h1>Мои цели</h1>
        <div className={styles.buttons}>
          <Button image={Plus} onClick={() => setIsOpenModal(true)} />
        </div>
      </section>
      <div className={styles.goals_container}>
        <Goal
          image={GoalCar}
          title="Машина"
          accumulated={300000}
          total={1500000}
          deadline={1825785465}
        />
        <Goal
          image={GoalFlat}
          title="Квартира"
          accumulated={12581528}
          total={100000000}
          deadline={1725785465}
        />
      </div>

      <CreateGoalModal
        isOpen={isOpenModal}
        onClose={() => setIsOpenModal(false)}
      />
    </div>
  );
}
