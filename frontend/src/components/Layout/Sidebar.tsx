const NAV_ITEMS = [
  { href: "/search",      label: "Search",        icon: "search" },
  { href: "/documents",   label: "Documents",     icon: "file-text" },
  { href: "/documents/upload", label: "Upload",   icon: "upload" },
  { href: "/chat",        label: "Chat",          icon: "message-circle" },
  { href: "/knowledge",   label: "Knowledge Base", icon: "database" },
] as const;

export function Sidebar() {
  return (
    <aside className="w-56 min-h-screen bg-surface-100 border-r border-surface-200 flex flex-col">
      <div className="p-4 border-b border-surface-200">
        <div className="text-xs font-semibold text-surface-600 uppercase tracking-wide">
          Navigation
        </div>
      </div>

      <nav className="flex-1 overflow-y-auto">
        {NAV_ITEMS.map((item) => (
          <a
            key={item.href}
            href={item.href}
            className="flex items-center gap-3 px-4 py-2.5 text-sm text-surface-700 hover:bg-surface-150 hover:text-surface-900 transition-colors"
          >
            <span className={`i-${item.icon} text-surface-400`} />
            <span>{item.label}</span>
          </a>
        ))}
      </nav>

      <div className="p-3 border-t border-surface-200 text-xs text-surface-400">
        Agentic RAG · Hybrid Search
      </div>
    </aside>
  );
}