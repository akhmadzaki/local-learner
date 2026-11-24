import { useState, useEffect } from "react";
import { documentsApi } from "../api/documents";
import type { Document, DocumentRaw } from "../types/document";

/** Convert snake_case backend document to camelCase frontend document. */
function toDocument(raw: DocumentRaw): Document {
  return {
    id: raw.id,
    title: raw.title,
    source: raw.source as Document["source"],
    status: raw.status as Document["status"],
    chunkCount: raw.chunk_count,
    metadata: raw.metadata,
    createdAt: raw.created_at,
    updatedAt: raw.updated_at,
  };
}

export function useDocuments() {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  async function fetchDocuments() {
    setLoading(true);
    setError(null);
    try {
      const res = await documentsApi.list();
      setDocuments((res.data ?? []).map(toDocument));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to fetch documents");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    fetchDocuments();
  }, []);

  return { documents, loading, error, refresh: fetchDocuments };
}