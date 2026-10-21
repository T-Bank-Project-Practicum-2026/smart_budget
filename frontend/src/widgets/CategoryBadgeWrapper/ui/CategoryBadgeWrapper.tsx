import React from "react";
import styles from "./CategoryBadgeWrapper.module.scss";
import {
  CategoryBadge,
  useCategories,
  CategoryBadgeSkeleton,
} from "../../../entities/category";
import { AddCategoryBadge } from "../../../features/addCategory/ui/AddCategoryBadge";

export function CategoryBadgeWrapper() {
  const { categories, isLoading, error } = useCategories();

  if (isLoading)
    return (
      <div className={styles.container}>
        {[0, 1, 2, 3, 4].map((c) => (
          <CategoryBadgeSkeleton key={c} />
        ))}
        <AddCategoryBadge />
      </div>
    );
  if (error)
    return (
      <div className={styles.container}>
        {[0, 1, 2, 3, 4].map((c) => (
          <CategoryBadgeSkeleton />
        ))}
        <AddCategoryBadge />
      </div>
    );

  return (
    <div className={styles.container}>
      {categories.map((c) => (
        <CategoryBadge
          key={c.id}
          title={c.title}
          color={c.color}
          value={c.value}
        />
      ))}
      <AddCategoryBadge />
    </div>
  );
}
