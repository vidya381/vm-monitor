export function DemoBanner() {
  if (process.env.NEXT_PUBLIC_DEMO_MODE !== "true") return null;

  return (
    <div className="w-full bg-accent/10 border-b border-accent/20 px-4 py-2 text-sm text-accent text-center shrink-0 flex flex-wrap items-center justify-center gap-x-1">
      <span>Demo mode — all data is simulated. Changes are disabled.</span>
      <a
        href="https://github.com/vidya381/vm-monitor"
        target="_blank"
        rel="noopener noreferrer"
        className="underline hover:opacity-80 transition-opacity whitespace-nowrap"
      >
        View on GitHub
      </a>
    </div>
  );
}
