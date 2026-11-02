import { GET } from "../client";

export type ImportError = {
  line: number;
  message: string;
};

export type Import = {
  conflictRows: number;
  createdAt: string;
  duplicateRows: number;
  errors: ImportError[];
  errorsTruncated: boolean;
  failedRows: number;
  fileHash: string;
  fileName: string;
  finishedAt: string;
  id: string;
  importedRows: number;
  skippedRows: number;
  status: string;
  totalRows: number;
};

export function imports() {
  return GET<Import[]>("/v1/imports");
}
