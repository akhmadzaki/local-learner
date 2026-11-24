import { useState, useEffect } from "react";
import { documentsApi } from "../../api/documents";
import type { Document } from "../../types/document";

export function DocumentList() {
  const [docs, setDocs] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    documentsApi.list()
      .then((res) => setDocs(res.documents))
      .catch((err) => {
        setError(err instanceof Error ? err.message : "Failed to load documents");
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <div className="text-surface-400 p-4">Loading documents…</div>;
  }

  return (
    <section>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-lg font-semibold text-surface-800">Documents</h1>
        <a
          href="/documents/upload"
          className="px-4 py-2 bg-primary-500 text-white rounded-lg text-sm"
        >
          Upload
        </a>
      </div>

      {error && (
        <div className="mb-4 px-4 py-3 rounded-lg bg-yellow-50 border border-yellow-200 text-yellow-800 text-sm">
          {error}
        </div>
      )}

      <div className="overflow-x-auto rounded-lg border border-surface-200">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-surface-100 text-surface-600 text-left">
              <th className="px-4 py-3 font-medium">Title</th>
              <th className="px-4 py-3 font-medium">Source</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Chunks</th>
              <th className="px-4 py-3 font-medium">Created</th>
              <th className="px-4 py-3 font-medium" />
            </tr>
          </thead>
          <tbody>
            {docs.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-surface-400">
                  No documents yet. Upload your first document.
                </td>
              </tr>
            )}
            {docs.map((doc) => (
              <tr key={doc.id} className="border-b border-surface-100 hover:bg-surface-50">
                <td className="px-4 py-3 font-medium text-surface-800">{doc.title}</td>
                <td className="px-4 py-3 text-surface-500">{doc.source}</td>
                <td className="px-4 py-3">
                  <StatusBadge status={doc.status} />
                </td>
                <td className="px-4 py-3 text-surface-600">{doc.chunkCount}</td>
                <td className="px-4 py-3 text-surface-400 text-xs">
                  {formatDate(doc.createdAt)}
                </td>
                <td className="px-4 py-3">
                  <button
                    onClick={() => documentsApi.delete(doc.id)}
                    className="text-red-500 hover:text-red-700 text-xs"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function StatusBadge({ status }: { status: Document["status"] }) {
  const colors: Record<string, string> = {
    pending: "bg-yellow-100 text-yellow-800",
    processing: "bg-blue-100 text-blue-800",
    indexed: "bg-green-100 text-green-800",
    failed: "bg-red-100 text-red-800",
  };
  return (
    <span className={`px-2 py-0.5 rounded text-xs font-medium ${colors[status] ?? "bg-surface-200 text-surface-600"}`}>
      {status}
    </span>
  );
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
}