import Link from "next/link";

export function Nav() {
  return (
    <header className="border-b border-border bg-background sticky top-0 z-10">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex h-14 items-center justify-between">
          <div className="flex items-center gap-6">
            <Link href="/" className="font-semibold text-foreground">
              VMMonitor
            </Link>
            <nav className="hidden sm:flex items-center gap-4 text-sm">
              <Link
                href="/"
                className="text-muted-foreground hover:text-foreground transition-colors"
              >
                Apps
              </Link>
              <Link
                href="/vms"
                className="text-muted-foreground hover:text-foreground transition-colors"
              >
                VMs
              </Link>
            </nav>
          </div>
        </div>
      </div>
    </header>
  );
}
