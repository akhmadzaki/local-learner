import { useState, useCallback } from "react";
import { searchApi } from "../api/search";
import type { SearchResponse, SearchMode } from "../types/search";

export function useSearch() {
  const [results, setResults] = useState<SearchResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const search = useCallback(async (query: string, mode: SearchMode = "hybrid") => {
    if (!query.trim()) return;
    setLoading(true);
    setError(null);
    try {
      const res = await searchApi.search({ query, mode, top_k: 10 });
      setResults(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Search failed");
    } finally {
      setLoading(false);
    }
  }, []);

  return { results, loading, error, search };
}