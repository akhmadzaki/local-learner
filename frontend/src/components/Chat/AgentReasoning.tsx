interface AgentReasoningProps {
  trace: string;
  confidence: number;
  open: boolean;
}

export function AgentReasoning({ trace, confidence, open }: AgentReasoningProps) {
  if (!trace) return null;

  return (
    <details open={open} className="mt-2">
      <summary className="text-xs text-surface-500 cursor-pointer">
        Reasoning trace · {(confidence * 100).toFixed(0)}% confidence
      </summary>
      <div className="mt-1 p-2 rounded bg-surface-50 text-xs text-surface-600 font-mono whitespace-pre-wrap">
        {trace}
      </div>
    </details>
  );
}