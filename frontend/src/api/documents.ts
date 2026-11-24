import { api, ApiError } from "./client";
import type {
  Document,
  DocumentListResponse,
  IngestDocumentRequest,
} from "../types/document";

const API_BASE = "/api/v1";

export const documentsApi = {
  list: (offset = 0, limit = 20) =>
    api.get<DocumentListResponse>(
      `/documents?offset=${offset}&limit=${limit}`,
    ),

  get: (id: string) =>
    api.get<Document>(`/documents/${id}`),

  ingest: (req: IngestDocumentRequest) =>
    api.post<Document>("/documents", req),

  /**
   * Upload a document file (PDF, Markdown, HTML, TXT).
   * Sends the file as multipart/form-data; the backend extracts text content.
   */
  upload: async (
    file: File,
    title?: string,
    source?: string,
    signal?: AbortSignal,
  ): Promise<Document> => {
    const formData = new FormData();
    formData.append("file", file);
    if (title) formData.append("title", title);
    if (source) formData.append("source", source);

    const res = await fetch(`${API_BASE}/documents/upload`, {
      method: "POST",
      body: formData,
      signal,
    });

    if (!res.ok) {
      const errBody = await res.json().catch(() => ({ message: res.statusText }));
      throw new ApiError(res.status, errBody.message ?? "Upload failed");
    }

    return res.json() as Promise<Document>;
  },

  delete: (id: string) =>
    api.delete<{ deleted: boolean }>(`/documents/${id}`),
};