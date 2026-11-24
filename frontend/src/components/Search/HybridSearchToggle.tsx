import type { SearchMode } from "../../types/search";

interface HybridSearchToggleProps {
  mode: SearchMode;
  onChange: (mode: SearchMode) => void;
}

const MODES: { value: SearchMode; label: string; desc: string }[] = [
  { value: "vector",  label: "Vector",  desc: "Semantic similarity" },
  { value: "keyword", label: "Keyword", desc: "Full-text / BM25" },
  { value: "hybrid",  label: "Hybrid",  desc: "RRF fusion" },
];

export function HybridSearchToggle({ mode, onChange }: HybridSearchToggleProps) {
  return (
    <div className="flex items-center gap-2 mb-4">
      <span className="text-xs text-surface-500 font-medium">Mode:</span>
      {MODES.map((m) => (
        <button
          key={m.value}
          onClick={() => onChange(m.value)}
          className={`px-3 py-1.5 rounded-full text-xs font-medium transition-colors ${
            mode === m.value
              ? "bg-primary-500 text-white"
              : "bg-surface-100 text-surface-600 hover:bg-surface-200"
          }`}
          title={m.desc}
        >
          {m.label}
        </button>
      ))}
    </div>
  );
}