import { RadialChart } from "../../../widgets/RadialChart/ui/RadialChart";
import Container from "../../../shared/ui/Container";
import { CategoryBadgeWrapper } from "../../../widgets/CategoryBadgeWrapper";
import ExpensesWrapper from "../../../shared/ui/ExpensesWrapper/ExpensesWrapper";
import styles from "./ButgetPage.module.scss";
import { RadialChartSkeleton } from "../../../widgets/RadialChart/ui/RadialChartSkeleton";
import ButgetActionButtons from "../../../widgets/butgetActionButtons/ui/ButgetActionButtons";

export function BudgetPage() {
  return (
    <div className={styles.page}>
      <section className={styles.section}>
        <h1>Бюджет на октябрь</h1>
        {/* <RadialChart /> */}
      </section>

      <section>
        <ButgetActionButtons />
      </section>

      <section className={styles.section}>
        <h2>Расходы по категориям в октябре</h2>
        {/* <CategoryBadgeWrapper /> */}
      </section>

      <section className={styles.section}>
        <ExpensesWrapper />
      </section>
    </div>
  );
}
