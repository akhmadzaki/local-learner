import type { ConversationTurn } from "../../types/agent";

interface MessageBubbleProps {
  turn: ConversationTurn;
}

export function MessageBubble({ turn }: MessageBubbleProps) {
  const isUser = turn.role === "user";

  return (
    <div className={`flex ${isUser ? "justify-end" : "justify-start"} fade-in`}>
      <div
        className={`max-w-[80%] p-3 rounded-2xl ${
          isUser
            ? "bg-primary-100 text-primary-900 ml-auto"
            : "bg-surface-100 text-surface-800"
        }`}
      >
        <div className="text-sm whitespace-pre-wrap">{turn.content}</div>

        {!isUser && turn.sources.length > 0 && (
          <div className="mt-2 pt-2 border-t border-surface-200">
            <details>
              <summary className="text-xs text-surface-500 cursor-pointer">
                {turn.sources.length} source{turn.sources.length !== 1 ? "s" : ""}
              </summary>
              <div className="mt-1 space-y-1">
                {turn.sources.map((s, i) => (
                  <div key={i} className="text-xs text-surface-400 truncate">
                    <span className="text-primary-500">[{s.score.toFixed(2)}]</span>{" "}
                    {s.documentTitle}
                  </div>
                ))}
              </div>
            </details>
          </div>
        )}

        <div className="text-xs text-surface-400 mt-1">
          {new Date(turn.createdAt).toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" })}
        </div>
      </div>
    </div>
  );
}