import { useState, useCallback } from "react";
import type { FormEvent } from 'react';

interface SearchBarProps {
  onSearch?: (query: string) => void;
  placeholder?: string;
}

export function SearchBar({ onSearch, placeholder = "Search the knowledge base…" }: SearchBarProps) {
  const [query, setQuery] = useState("");

  const handleSubmit = useCallback((e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (query.trim() && onSearch) {
      onSearch(query.trim());
    }
  }, [query, onSearch]);

  return (
    <form onSubmit={handleSubmit} className="relative">
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        className="w-full px-4 py-2 rounded-full border border-surface-200 bg-surface-50 text-sm text-surface-800 placeholder-surface-400 outline-none focus:border-primary-300 focus:ring-2 focus:ring-primary-200"
        placeholder={placeholder}
      />
      <button
        type="submit"
        className="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 text-surface-400 hover:text-primary-500"
      >
        <span className="i-search" />
      </button>
    </form>
  );
}