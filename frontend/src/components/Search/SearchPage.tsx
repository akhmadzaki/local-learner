import { useState, useCallback } from "react";
import { searchApi } from "../../api/search";
import { SearchBar } from "./SearchBar";
import { SearchResults } from "./SearchResults";
import { HybridSearchToggle } from "./HybridSearchToggle";
import type { SearchResponse, SearchMode } from "../../types/search";

export function SearchPage() {
  const [results, setResults] = useState<SearchResponse | null>(null);
  const [mode, setMode] = useState<SearchMode>("hybrid");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSearch = useCallback(async (query: string) => {
    setLoading(true);
    setError("");
    try {
      const res = await searchApi.search({ query, mode, topK: 10 });
      setResults(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Search failed");
      setResults(null);
    } finally {
      setLoading(false);
    }
  }, [mode]);

  return (
    <section>
      <div className="max-w-4xl mx-auto">
        <div className="mb-6">
          <SearchBar onSearch={handleSearch} />
        </div>

        <HybridSearchToggle
          mode={mode}
          onChange={setMode}
        />

        {loading && (
          <div className="text-surface-400 text-sm py-4">Searching…</div>
        )}

        {error && (
          <div className="bg-red-50 text-red-800 p-3 rounded-lg text-sm">{error}</div>
        )}

        {results && (
          <div className="mt-4">
            <div className="text-xs text-surface-400 mb-2">
              {results.totalHits} results in {results.queryTimeMs}ms · {results.mode} mode
            </div>
            <SearchResults results={results.results} />
          </div>
        )}
      </div>
    </section>
  );
}