import { useState, useRef } from "react";
import { documentsApi } from "../../api/documents";
import type { FormEvent } from 'react';

export function DocumentUpload() {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [source, setSource] = useState("upload");
  const [status, setStatus] = useState<"idle" | "uploading" | "success" | "error">("idle");
  const [message, setMessage] = useState("");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploadMode, setUploadMode] = useState<"text" | "file">("text");
  const fileInputRef = useRef<HTMLInputElement>(null);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (uploadMode === "file") {
      if (!title.trim() || !selectedFile) return;

      setStatus("uploading");
      try {
        await documentsApi.upload(selectedFile, title, source);
        setStatus("success");
        setMessage(`"${title}" ingested successfully.`);
        setTitle("");
        setSelectedFile(null);
        setContent("");
        if (fileInputRef.current) fileInputRef.current.value = "";
      } catch (err) {
        setStatus("error");
        setMessage(err instanceof Error ? err.message : "Upload failed");
      }
    } else {
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
  }

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0] ?? null;
    setSelectedFile(file);
    // Auto-fill title from filename if not already set
    if (file && !title) {
      setTitle(file.name);
    }
  }

  function switchMode(mode: "text" | "file") {
    setUploadMode(mode);
    setStatus("idle");
    setMessage("");
  }

  return (
    <section className="max-w-2xl mx-auto">
      <h1 className="text-lg font-semibold text-surface-800 mb-4">Upload Document</h1>

      {/* Mode toggle */}
      <div className="flex gap-2 mb-4">
        <button
          type="button"
          onClick={() => switchMode("text")}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            uploadMode === "text"
              ? "bg-primary-500 text-white"
              : "bg-surface-100 text-surface-600 hover:bg-surface-200"
          }`}
        >
          Paste Text
        </button>
        <button
          type="button"
          onClick={() => switchMode("file")}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            uploadMode === "file"
              ? "bg-primary-500 text-white"
              : "bg-surface-100 text-surface-600 hover:bg-surface-200"
          }`}
        >
          Upload File
        </button>
      </div>

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

        {uploadMode === "file" ? (
          <div>
            <label className="block text-sm font-medium text-surface-600 mb-1">
              File (PDF, Markdown, HTML, TXT)
            </label>
            <div className="flex items-center gap-3">
              <input
                ref={fileInputRef}
                type="file"
                accept=".pdf,.md,.markdown,.html,.htm,.txt"
                onChange={handleFileChange}
                className="block w-full text-sm text-surface-500 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-medium file:bg-primary-50 file:text-primary-700 hover:file:bg-primary-100"
              />
            </div>
            {selectedFile && (
              <p className="mt-2 text-xs text-surface-500">
                {selectedFile.name} ({(selectedFile.size / 1024).toFixed(1)} KB)
              </p>
            )}
          </div>
        ) : (
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
        )}

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