import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Layout } from "./components/Layout/Layout";
import { DocumentList } from "./components/Documents/DocumentList";
import { DocumentUpload } from "./components/Documents/DocumentUpload";
import { SearchPage } from "./components/Search/SearchPage";
import { ChatInterface } from "./components/Chat/ChatInterface";
import { KnowledgeBaseStats } from "./components/KnowledgeBase/KnowledgeBaseStats";

export function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<Navigate to="/search" />} />
          <Route path="/search" element={<SearchPage />} />
          <Route path="/documents" element={<DocumentList />} />
          <Route path="/documents/upload" element={<DocumentUpload />} />
          <Route path="/chat" element={<ChatInterface />} />
          <Route path="/chat/:id" element={<ChatInterface />} />
          <Route path="/knowledge" element={<KnowledgeBaseStats />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
}