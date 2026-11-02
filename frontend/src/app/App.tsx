import { useState } from "react";
import "./styles";
import { Header } from "../widgets/header";
import { BudgetPage } from "../pages/budget";
import Footer from "../shared/ui/Footer";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { GoalsPage } from "../pages/goals";
import { AuthPage, RegisterPage } from "../pages/sign-in";
import { ProfilePage } from "../pages/profile";
import LandingPage from "../pages/landing";
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "../shared/api/queryClient";

function App() {
  return (
    <>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Header />
          <Routes>
            <Route path="/landing" element={<LandingPage />} />
            <Route path="/budget" element={<BudgetPage />} />
            <Route path="/goals" element={<GoalsPage />} />
            <Route path="/login" element={<AuthPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/profile" element={<ProfilePage />} />
          </Routes>
          <Footer />
        </BrowserRouter>
      </QueryClientProvider>
    </>
  );
}

export default App;
