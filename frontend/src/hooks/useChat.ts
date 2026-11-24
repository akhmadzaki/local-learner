import { useState, useCallback } from "react";
import { chatApi } from "../api/search";
import type { ChatResponse, AgentMode, ConversationTurn } from "../types/agent";
import type { ApiResponse } from "../types/search";

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
      const res = await chatApi.send({ message: text, conversation_id: convId ?? undefined, mode }) as unknown as ApiResponse<ChatResponse>;
      const data = res.data;
      setConvId(data.conversation_id);

      const assistantTurn: ConversationTurn = {
        id: crypto.randomUUID(),
        conversationId: data.conversation_id,
        role: "assistant",
        content: data.answer,
        sources: data.sources ?? [],
        tokenUsage: data.token_usage ?? {},
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