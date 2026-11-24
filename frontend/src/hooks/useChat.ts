import { useState, useCallback } from "react";
import { chatApi } from "../api/search";
import type { ChatResponse, AgentMode, ConversationTurn } from "../types/agent";

export function useChat() {
  const [messages, setMessages] = useState<ConversationTurn[]>([]);
  const [convId, setConvId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const send = useCallback(async (text: string, mode: AgentMode = "reasoned") => {
    if (!text.trim() || loading) return;

    setLoading(true);
    setError(null);

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
      const res = await chatApi.send({ message: text, conversationId: convId ?? undefined, mode });
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
      setError(err instanceof Error ? err.message : "Chat request failed");
    } finally {
      setLoading(false);
    }
  }, [convId, loading]);

  return { messages, convId, loading, error, send };
}