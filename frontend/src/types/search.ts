export type SearchMode = "vector" | "keyword" | "hybrid";

export interface SearchRequest {
  query: string;
  mode: SearchMode;
  topK?: number;
  minScore?: number;
  hybridAlpha?: number;
  filters?: Record<string, unknown>;
}

export interface SearchResult {
  chunkId: string;
  documentId: string;
  documentTitle: string;
  content: string;
  score: number;
  scoreVector: number;
  scoreKeyword: number;
}

export interface SearchResponse {
  results: SearchResult[];
  totalHits: number;
  queryTimeMs: number;
  mode: SearchMode;
}