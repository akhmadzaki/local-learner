export function Header() {
  return (
    <header className="flex items-center justify-between h-14 px-6 bg-white border-b border-surface-200">
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary-500 to-primary-700" />
        <span className="font-semibold text-surface-800 text-sm">
          Knowledge Base
        </span>
      </div>

      <div className="flex items-center gap-2 text-xs text-surface-500">
        <span className="i-more-horizontal" />
        <span>v1.0</span>
      </div>
    </header>
  );
}