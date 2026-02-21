import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { AppCard } from "../app-card";
import { App } from "@/lib/types";

// next/link renders a plain <a> in test environments
vi.mock("next/link", () => ({
  default: ({ href, children }: { href: string; children: React.ReactNode }) => (
    <a href={href}>{children}</a>
  ),
}));

const baseApp: App = {
  id: "abc-123",
  vm_id: "vm-1",
  vm_name: "oracle-amd-vm1",
  name: "spendo",
  type: "systemd",
  last_status: "running",
  created_at: "2024-01-01T00:00:00Z",
};

describe("AppCard", () => {
  it("renders app name", () => {
    render(<AppCard app={baseApp} />);
    expect(screen.getByText("spendo")).toBeInTheDocument();
  });

  it("renders VM name", () => {
    render(<AppCard app={baseApp} />);
    expect(screen.getByText("oracle-amd-vm1")).toBeInTheDocument();
  });

  it("renders app type", () => {
    render(<AppCard app={baseApp} />);
    expect(screen.getByText("systemd")).toBeInTheDocument();
  });

  it("links to the app detail page", () => {
    render(<AppCard app={baseApp} />);
    const link = screen.getByRole("link");
    expect(link).toHaveAttribute("href", "/apps/abc-123");
  });

  it("shows status badge", () => {
    render(<AppCard app={baseApp} />);
    expect(screen.getByText("Running")).toBeInTheDocument();
  });

  it("shows never for missing last_checked_at", () => {
    render(<AppCard app={{ ...baseApp, last_checked_at: undefined }} />);
    expect(screen.getByText(/never/)).toBeInTheDocument();
  });
});
