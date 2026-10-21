import React from "react";
import styles from "./CategoryBadge.module.scss";

export function CategoryBadgeSkeleton() {
  return <div className={`${styles.skeleton_container}`} />;
}
