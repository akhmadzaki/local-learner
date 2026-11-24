import { api } from "./client";
import type {
  Document,
  DocumentListResponse,
  IngestDocumentRequest,
} from "../types/document";

export const documentsApi = {
  list: (offset = 0, limit = 20) =>
    api.get<DocumentListResponse>(
      `/documents?offset=${offset}&limit=${limit}`,
    ),

  get: (id: string) =>
    api.get<Document>(`/documents/${id}`),

  ingest: (req: IngestDocumentRequest) =>
    api.post<Document>("/documents", req),

  delete: (id: string) =>
    api.delete<{ deleted: boolean }>(`/documents/${id}`),
};