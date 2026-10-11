import RadialChart from "../../shared/ui/RadialChart";
import Container from "../../shared/ui/Container";
import CategoryBadgeWrapper from "../../shared/ui/CategoryBadgeWrapper";
import ExpensesWrapper from "../../shared/ui/ExpensesWrapper/ExpensesWrapper";

export default function BudgetPage() {
  return (
    <>
      <Container>
        <h1>Бюджет на октябрь</h1>
      </Container>
      <RadialChart />
      <Container>
        <h2>Расходы по категориям в октябре</h2>
        <CategoryBadgeWrapper />
        <ExpensesWrapper />
      </Container>
    </>
  );
}
