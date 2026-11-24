export interface Document {
  id: string;
  title: string;
  source: "upload" | "web" | "api" | "scrape";
  status: "pending" | "processing" | "indexed" | "failed";
  /** Mapped from backend snake_case chunk_count */
  chunkCount: number;
  metadata: Record<string, unknown>;
  /** Mapped from backend snake_case created_at */
  createdAt: string;
  /** Mapped from backend snake_case updated_at */
  updatedAt: string;
}

/** Raw shape returned by the backend (snake_case JSON). */
export interface DocumentRaw {
  id: string;
  title: string;
  source: string;
  status: string;
  chunk_count: number;
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

/** The document list endpoint returns { status, data } where data is DocumentRaw[]. */
export type DocumentListResponse = {
  status: string;
  data: DocumentRaw[];
};

export interface IngestDocumentRequest {
  title: string;
  content: string;
  source: string;
}

export type DocumentStatus = Document["status"];
export type DocumentSource = Document["source"];