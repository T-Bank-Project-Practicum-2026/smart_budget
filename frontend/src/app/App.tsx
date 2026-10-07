import { useState } from "react";
import "./styles";
import Header from "../widgets/ui/header";
import BudgetPage from "../pages/BudgetPage";
import Footer from "../shared/ui/Footer";

function App() {
  return (
    <>
      <Header />
      <BudgetPage />
      <Footer />
    </>
  );
}

export default App;
