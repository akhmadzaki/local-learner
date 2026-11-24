export interface Document {
  id: string;
  title: string;
  source: "upload" | "web" | "api" | "scrape";
  status: "pending" | "processing" | "indexed" | "failed";
  chunkCount: number;
  metadata: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface DocumentListResponse {
  documents: Document[];
  total: number;
  offset: number;
  limit: number;
}

export interface IngestDocumentRequest {
  title: string;
  content: string;
  source: string;
}

export type DocumentStatus = Document["status"];
export type DocumentSource = Document["source"];