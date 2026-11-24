import type { SearchResult } from "./search";

export type AgentMode = "direct" | "reasoned" | "deep";

export interface ChatRequest {
  message: string;
  conversationId?: string;
  mode?: AgentMode;
  streamResponse?: boolean;
}

export interface ChatResponse {
  answer: string;
  sources: SearchResult[];
  reasoningTrace: string;
  confidence: number;
  conversationId: string;
  tokenUsage: Record<string, number>;
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