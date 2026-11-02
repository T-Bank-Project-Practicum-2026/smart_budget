import { POST } from "../client";
import type { Import } from "./imports";

export function importTbank(file: File) {
  const formData = new FormData();
  formData.append("file", file);

  return POST<Import>("/v1/imports/tbank", formData);
}
