import "./styles";

import { BudgetPage } from "@/pages/budget";
import { LandingPage } from "@/pages/landing";
import { GoalsPage } from "@/pages/goals";
import { LoginPage, RegisterPage } from "@/pages/sign-in";
import { ProfilePage } from "@/pages/profile";

import { Header } from "@widgets/header";
import { Footer } from "@/widgets/footer";

import { BrowserRouter, Route, Routes } from "react-router-dom";
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "@shared/api/queryClient";

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
            <Route path="/login" element={<LoginPage />} />
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
