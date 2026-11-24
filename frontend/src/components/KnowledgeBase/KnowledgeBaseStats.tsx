import { useState, useEffect } from "react";
import { api } from "../../api/client";

interface IndexStats {
  total_documents: number;
  total_chunks: number;
  total_vectors: number;
  last_indexed_at: string;
  index_health: "healthy" | "degraded" | "unhealthy";
}

interface IndexStatsResponse {
  data: IndexStats;
  status: string;
}

export function KnowledgeBaseStats() {
  const [stats, setStats] = useState<IndexStatsResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.get<IndexStatsResponse>("/index/stats")
      .then(setStats)
      .finally(() => setLoading(false));
  }, []);

  return (
    <section>
      <h1 className="text-lg font-semibold text-surface-800 mb-4">Knowledge Base</h1>

      {loading && <div className="text-surface-400 text-sm">Loading stats…</div>}

      {stats && (
        <div className="grid grid-cols-2 gap-4 mb-6">
          <StatCard label="Documents" value={stats.data.total_documents} />
          <StatCard label="Chunks" value={stats.data.total_chunks} />
          <StatCard label="Vectors" value={stats.data.total_vectors} />
          <StatCard
            label="Index Health"
            value={stats.data.index_health}
            color={stats.data.index_health === "healthy" ? "text-green-600" : "text-red-600"}
          />
        </div>
      )}

      <div className="mt-4">
        <button
          onClick={() => api.post("/index/build", {})}
          className="px-4 py-2 bg-primary-500 text-white rounded-lg text-sm"
        >
          Rebuild Index
        </button>
      </div>
    </section>
  );
}

function StatCard({ label, value, color = "text-surface-800" }: { label: string; value: string | number; color?: string }) {
  return (
    <div className="p-4 rounded-lg border border-surface-200 bg-white">
      <div className="text-xs text-surface-500 uppercase tracking-wide">{label}</div>
      <div className={`text-2xl font-semibold mt-1 ${color}`}>{value}</div>
    </div>
  );
}