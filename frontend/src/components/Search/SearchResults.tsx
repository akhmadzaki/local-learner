import type { SearchResult } from "../../types/search";

interface SearchResultsProps {
  results: SearchResult[];
}

export function SearchResults({ results }: SearchResultsProps) {
  if (!results || results.length === 0) {
    return (
      <div className="text-center py-8 text-surface-400 text-sm">
        No results found. Try a different query or mode.
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {results.map((r, i) => (
        <div
          key={r.chunk_id}
          className="p-4 rounded-lg border border-surface-200 bg-white hover:border-primary-200 transition-colors"
        >
          <div className="flex items-center justify-between mb-1">
            <span className="text-xs font-medium text-primary-600">
              {r.document_title}
            </span>
            <span className="text-xs text-surface-400">
              Score: {(r.score * 100).toFixed(1)}%
            </span>
          </div>

          <p className="text-sm text-surface-700 leading-relaxed line-clamp-4">
            {r.content}
          </p>

          <div className="flex items-center gap-2 mt-2 text-xs text-surface-400">
            {r.score_vector > 0 && (
              <span className="px-1.5 py-0.5 rounded bg-blue-50 text-blue-700">
                vector: {(r.score_vector * 100).toFixed(0)}%
              </span>
            )}
            {r.score_keyword > 0 && (
              <span className="px-1.5 py-0.5 rounded bg-green-50 text-green-700">
                keyword: {(r.score_keyword * 100).toFixed(0)}%
              </span>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}