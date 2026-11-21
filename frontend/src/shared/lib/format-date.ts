export function formatDate(dateStr: string): string {
  const d = new Date(dateStr);

  const time = d.toLocaleTimeString("ru-RU");
  const date = d.toLocaleDateString("ru-RU");

  return `${time} ${date}`;
}
