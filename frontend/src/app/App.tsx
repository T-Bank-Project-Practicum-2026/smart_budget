import { useState } from "react";
import "./styles";
import Header from "../widgets/ui/header";
import BudgetPage from "../pages/BudgetPage";
import Footer from "../shared/ui/Footer";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import GoalsPage from "../pages/GoalsPage";

function App() {
  return (
    <>
      <BrowserRouter>
        <Header />
        <Routes>
          <Route path="/butget" element={<BudgetPage />} />
          <Route path="/goals" element={<GoalsPage />} />
        </Routes>
        <Footer />
      </BrowserRouter>
    </>
  );
}

export default App;
