import { useNavigate } from "react-router-dom";
import Button from "../../../shared/ui/Button";
import HeaderButton from "../../../shared/ui/HeaderButton";
import styles from "./LandingPage.module.scss";
import LandingImage from "@shared/ui/logo/landing_image.png";

export function LandingPage() {
  const navigate = useNavigate();

  return (
    <div className={styles.page}>
      <div className={styles.title_section}>
        <div className={styles.title}>
          <div className={styles.title_text}>
            <h1>Умный бюджет</h1>
            <h2>
              Удобное web-приложение для ведения личного финансового учета
            </h2>
          </div>
          <div className={styles.buttons}>
            <Button onClick={() => navigate("/register", { replace: true })}>
              Начать
            </Button>
          </div>
        </div>
        <img src={LandingImage} />
      </div>
    </div>
  );
}
