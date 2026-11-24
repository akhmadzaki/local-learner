import type { SearchResult } from "./search";

export type AgentMode = "direct" | "reasoned" | "deep";

export interface ChatRequest {
  message: string;
  conversation_id?: string;
  mode?: AgentMode;
  stream_response?: boolean;
}

export interface ChatResponse {
  answer: string;
  sources: SearchResult[];
  reasoning_trace: string;
  confidence: number;
  conversation_id: string;
  token_usage: Record<string, number>;
}

export interface Conversation {
  id: string;
  title: string;
  turnCount: number;
  status: "active" | "archived";
  createdAt: string;
  updatedAt: string;
}

export interface ConversationTurn {
  id: string;
  conversationId: string;
  role: "user" | "assistant" | "system";
  content: string;
  sources: SearchResult[];
  tokenUsage: Record<string, number>;
  createdAt: string;
}