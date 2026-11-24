import { useState, useCallback } from "react";
import { chatApi } from "../../api/search";
import { MessageBubble } from "./MessageBubble";
import { AgentReasoning } from "./AgentReasoning";
import type { ChatResponse, AgentMode, ConversationTurn } from "../../types/agent";

export function ChatInterface() {
  const [messages, setMessages] = useState<ConversationTurn[]>([]);
  const [input, setInput] = useState("");
  const [convId, setConvId] = useState<string | null>(null);
  const [mode, setMode] = useState<AgentMode>("reasoned");
  const [loading, setLoading] = useState(false);

  const handleSend = useCallback(async () => {
    const text = input.trim();
    if (!text || loading) return;

    setInput("");
    setLoading(true);

    // Optimistic user message
    const userTurn: ConversationTurn = {
      id: crypto.randomUUID(),
      conversationId: convId ?? "",
      role: "user",
      content: text,
      sources: [],
      tokenUsage: {},
      createdAt: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, userTurn]);

    try {
      const res = await chatApi.send({
        message: text,
        conversationId: convId ?? undefined,
        mode,
      });

      setConvId(res.conversationId);

      const assistantTurn: ConversationTurn = {
        id: crypto.randomUUID(),
        conversationId: res.conversationId,
        role: "assistant",
        content: res.answer,
        sources: res.sources,
        tokenUsage: res.tokenUsage,
        createdAt: new Date().toISOString(),
      };
      setMessages((prev) => [...prev, assistantTurn]);
    } catch (err) {
      const errorTurn: ConversationTurn = {
        id: crypto.randomUUID(),
        conversationId: convId ?? "",
        role: "assistant",
        content: `Error: ${err instanceof Error ? err.message : "Request failed"}`,
        sources: [],
        tokenUsage: {},
        createdAt: new Date().toISOString(),
      };
      setMessages((prev) => [...prev, errorTurn]);
    } finally {
      setLoading(false);
    }
  }, [input, convId, mode, loading]);

  return (
    <section className="flex flex-col h-full">
      {/* Mode selector */}
      <div className="flex items-center gap-2 mb-3">
        <span className="text-xs text-surface-500">Reasoning:</span>
        {["direct", "reasoned", "deep"].map((m) => (
          <button
            key={m}
            onClick={() => setMode(m as AgentMode)}
            className={`px-3 py-1 rounded-full text-xs ${
              mode === m ? "bg-primary-500 text-white" : "bg-surface-100 text-surface-600"
            }`}
          >
            {m}
          </button>
        ))}
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto space-y-4 px-2">
        {messages.length === 0 && (
          <div className="text-center py-8 text-surface-400 text-sm">
            Ask a question to start a conversation with the agent.
          </div>
        )}
        {messages.map((msg) => (
          <MessageBubble key={msg.id} turn={msg} />
        ))}
        {loading && (
          <div className="text-surface-400 text-sm italic">Agent is thinking…</div>
        )}
      </div>

      {/* Input */}
      <div className="flex items-center gap-2 p-3 border-t border-surface-200 bg-white">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSend()}
          className="flex-1 px-4 py-2.5 rounded-full border border-surface-200 text-sm"
          placeholder="Ask the knowledge base…"
        />
        <button
          onClick={handleSend}
          disabled={loading}
          className="px-4 py-2.5 bg-primary-500 text-white rounded-full text-sm"
        >
          Send
        </button>
      </div>
    </section>
  );
}