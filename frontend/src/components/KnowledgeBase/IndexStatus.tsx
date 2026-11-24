interface IndexStatusProps {
  status: "idle" | "indexing" | "completed" | "failed";
  progress?: number; // 0–100
  message?: string;
}

export function IndexStatus({ status, progress, message }: IndexStatusProps) {
  const colors: Record<string, string> = {
    idle: "bg-surface-200 text-surface-500",
    indexing: "bg-blue-100 text-blue-800",
    completed: "bg-green-100 text-green-800",
    failed: "bg-red-100 text-red-800",
  };

  return (
    <div className="p-3 rounded-lg border border-surface-200 bg-white">
      <div className="flex items-center justify-between">
        <span className={`px-2 py-0.5 rounded text-xs font-medium ${colors[status]}`}>
          {status}
        </span>
        {progress !== undefined && (
          <span className="text-xs text-surface-500">{progress}%</span>
        )}
      </div>

      {progress !== undefined && (
        <div className="mt-2 h-1.5 rounded-full bg-surface-200 overflow-hidden">
          <div
            className="h-full bg-primary-500 transition-all"
            style={{ width: `${progress}%` }}
          />
        </div>
      )}

      {message && (
        <div className="mt-1 text-xs text-surface-400">{message}</div>
      )}
    </div>
  );
}