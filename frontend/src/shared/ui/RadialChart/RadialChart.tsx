import React from "react";
import styles from "./RadialChart.module.scss";
import { formatCurrency } from "../../lib/format-currency";

interface ChartSegment {
  label: string;
  value: number;
  plan: number;
  color: string;
}

interface RadialChartProps {
  data?: ChartSegment[];
  size?: number;
  thickness?: number;
}

const DEFAULT_DATA: ChartSegment[] = [
  { label: "Жилье", value: 40000, plan: 45000, color: "#61dafb" },
  { label: "Еда", value: 25000, plan: 30000, color: "#42b883" },
  { label: "Транспорт", value: 3000, plan: 5000, color: "#dd0031" },
  { label: "Шоппинг", value: 15000, plan: 12000, color: "#ff3e00" },
];

export default function RadialChart({
  data = DEFAULT_DATA,
  size = 240,
  thickness = 32,
}: RadialChartProps) {
  const total = data.reduce((sum, item) => sum + item.value, 0);
  const radius = (size - thickness) / 2;
  const center = size / 2;
  const circumference = 2 * Math.PI * radius;

  const gapPx = 2;

  let running = 0;
  const segments = data.map((segment) => {
    const fraction = segment.value / total;
    const dash = Math.max(0, fraction * circumference - gapPx);
    const gap = circumference - dash;

    const startAngle = (running / total) * 360 - 90;

    running += segment.value;

    return {
      ...segment,
      dash,
      gap,
      startAngle,
    };
  });

  const renderOrder = [...segments].reverse();

  return (
    <div className={styles.layout}>
      <div className={styles.left_skeleton} />
      <div className={styles.wrapper}>
        <div className={styles.chart} style={{ width: size, height: size }}>
          <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
            {/* Фоновое кольцо */}
            <circle
              className={styles.track}
              cx={center}
              cy={center}
              r={radius}
              strokeWidth={thickness}
            />

            {renderOrder.map((segment) => {
              const originalIndex = segments.indexOf(segment);
              return (
                <circle
                  key={originalIndex}
                  className={styles.segment}
                  cx={center}
                  cy={center}
                  r={radius}
                  strokeWidth={thickness}
                  strokeDasharray={`${segment.dash} ${segment.gap}`}
                  strokeDashoffset={0}
                  transform={`rotate(${segment.startAngle} ${center} ${center})`}
                  style={{
                    stroke: segment.color,
                    fill: "none",
                    animationDelay: `${originalIndex * 0.15}s`,
                  }}
                />
              );
            })}
          </svg>

          <div className={styles.center}>
            <span className={styles.total}>{formatCurrency(total)}</span>
          </div>
        </div>

        <table className={styles.table}>
          <thead>
            <tr>
              <th className={styles.thCategory}>Категории</th>
              <th className={styles.thValue}>Факт</th>
              <th className={styles.thValue}>План</th>
            </tr>
          </thead>
          <tbody>
            {data.map((segment, index) => {
              const percent = ((segment.value / total) * 100).toFixed(1);
              return (
                <tr key={index} className={styles.row}>
                  <td className={styles.tdCategory}>
                    <span
                      className={styles.dot}
                      style={{
                        backgroundColor: segment.color,
                      }}
                    />
                    <span className={styles.label}>{segment.label}</span>
                  </td>
                  <td className={styles.tdValue}>
                    {formatCurrency(segment.value)}
                  </td>
                  <td className={styles.tdValue}>
                    {formatCurrency(segment.plan)}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
      <div className={styles.right_skeleton} />
    </div>
  );
}
