export type SearchMode = "vector" | "keyword" | "hybrid";

export interface SearchRequest {
  query: string;
  mode: SearchMode;
  top_k?: number;
  min_score?: number;
  hybrid_alpha?: number;
  filters?: Record<string, unknown>;
}

export interface SearchResult {
  chunk_id: string;
  document_id: string;
  document_title: string;
  content: string;
  score: number;
  score_vector: number;
  score_keyword: number;
}

/** The shape returned directly by the Go backend (wrapped in ApiResponse.data). */
export interface SearchResponseData {
  results: SearchResult[];
  total_hits: number;
  query_time_ms: number;
  mode: SearchMode;
}

/** Standard API envelope used by helpers.RespondSuccess(). */
export interface ApiResponse<T> {
  status: string;
  data: T;
}

/** Convenience alias for the search endpoint response. */
export type SearchResponse = ApiResponse<SearchResponseData>;