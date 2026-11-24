import { useState, useEffect } from "react";
import { documentsApi } from "../api/documents";
import type { Document } from "../types/document";

export function useDocuments() {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  async function fetchDocuments() {
    setLoading(true);
    setError(null);
    try {
      const res = await documentsApi.list();
      setDocuments(res.documents);
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