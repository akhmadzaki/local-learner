import { useState } from "react";
import { documentsApi } from "../../api/documents";
import type { FormEvent } from 'react';

export function DocumentUpload() {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [source, setSource] = useState("upload");
  const [status, setStatus] = useState<"idle" | "uploading" | "success" | "error">("idle");
  const [message, setMessage] = useState("");

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!title.trim() || !content.trim()) return;

    setStatus("uploading");
    try {
      await documentsApi.ingest({ title, content, source });
      setStatus("success");
      setMessage(`"${title}" ingested successfully.`);
      setTitle("");
      setContent("");
    } catch (err) {
      setStatus("error");
      setMessage(err instanceof Error ? err.message : "Upload failed");
    }
  }

  return (
    <section className="max-w-2xl mx-auto">
      <h1 className="text-lg font-semibold text-surface-800 mb-4">Upload Document</h1>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-surface-600 mb-1">Title</label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="w-full px-3 py-2 rounded-lg border border-surface-200 bg-white text-sm"
            placeholder="Document title"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-surface-600 mb-1">Source</label>
          <select
            value={source}
            onChange={(e) => setSource(e.target.value)}
            className="w-full px-3 py-2 rounded-lg border border-surface-200 bg-white text-sm"
          >
            <option value="upload">Upload</option>
            <option value="web">Web Scrape</option>
            <option value="api">API</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-surface-600 mb-1">Content</label>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            className="w-full px-3 py-2 rounded-lg border border-surface-200 bg-white text-sm font-mono"
            rows={16}
            placeholder="Paste or type document content here…"
            required
          />
        </div>

        <button
          type="submit"
          disabled={status === "uploading"}
          className="px-6 py-2.5 bg-primary-500 text-white rounded-lg text-sm font-medium disabled:opacity-50"
        >
          {status === "uploading" ? "Ingesting…" : "Ingest Document"}
        </button>

        {message && (
          <div className={`p-3 rounded-lg text-sm ${status === "success" ? "bg-green-50 text-green-800" : "bg-red-50 text-red-800"}`}>
            {message}
          </div>
        )}
      </form>
    </section>
  );
}