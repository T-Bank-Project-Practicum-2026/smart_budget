import React from "react";
import styles from "./CategoryBadgeWrapper.module.scss";
import CategoryBadge from "../CategoryBadge/CategoryBadge";

export default function CategoryBadgeWrapper() {
  return (
    <div className={styles.container}>
      <CategoryBadge title="Супермаркеты" value={15000} color="#33ff00" />
      <CategoryBadge title="Кафе и рестораны" value={4250} color="#ff5733" />
      <CategoryBadge title="Транспорт" value={1200} color="#3399ff" />
      <CategoryBadge title="Развлечения" value={8750} color="#cc33ff" />
      <CategoryBadge title="Здоровье" value={3200} color="#ffcc00" />
      <CategoryBadge title="Одежда" value={15900} color="#ff3399" />
      <CategoryBadge title="Подписки" value={999} color="#00cccc" />
      <CategoryBadge title="Добавить категорию" color="#d1d1d1" />
    </div>
  );
}
