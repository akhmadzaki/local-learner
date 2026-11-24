import { api } from "./client";
import type { SearchRequest, SearchResponse } from "../types/search";
import type { ChatRequest, ChatResponse, Conversation } from "../types/agent";

export const searchApi = {
  search: (req: SearchRequest) =>
    api.post<SearchResponse>("/search", req),

  vectorSearch: (query: string, topK = 10) =>
    api.post<SearchResponse>("/search/vector", { query, topK }),

  keywordSearch: (query: string, topK = 10) =>
    api.post<SearchResponse>("/search/keyword", { query, topK }),

  hybridSearch: (query: string, topK = 10, alpha = 0.5) =>
    api.post<SearchResponse>("/search/hybrid", { query, topK, hybridAlpha: alpha }),
};

export const chatApi = {
  send: (req: ChatRequest) =>
    api.post<ChatResponse>("/chat", req),

  continue: (convId: string, req: ChatRequest) =>
    api.post<ChatResponse>(`/chat/${convId}`, req),

  getConversation: (convId: string) =>
    api.get<Conversation>(`/chat/${convId}`),

  listConversations: () =>
    api.get<Conversation[]>("/conversations"),

  deleteConversation: (convId: string) =>
    api.delete<{ deleted: boolean }>(`/chat/${convId}`),
};