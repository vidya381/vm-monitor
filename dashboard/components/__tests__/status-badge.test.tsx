import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { AppStatusBadge, VMStatusBadge } from "../status-badge";

describe("AppStatusBadge", () => {
  it("shows Running label for running status", () => {
    render(<AppStatusBadge status="running" />);
    expect(screen.getByText("Running")).toBeInTheDocument();
  });

  it("shows Stopped label for stopped status", () => {
    render(<AppStatusBadge status="stopped" />);
    expect(screen.getByText("Stopped")).toBeInTheDocument();
  });

  it("shows Unhealthy label for unhealthy status", () => {
    render(<AppStatusBadge status="unhealthy" />);
    expect(screen.getByText("Unhealthy")).toBeInTheDocument();
  });

  it("shows Unknown for empty/unrecognised status", () => {
    render(<AppStatusBadge status="" />);
    expect(screen.getByText("Unknown")).toBeInTheDocument();
  });

  it("running dot has animate-pulse class", () => {
    const { container } = render(<AppStatusBadge status="running" />);
    const dot = container.querySelector(".animate-pulse");
    expect(dot).not.toBeNull();
  });

  it("stopped dot does not have animate-pulse class", () => {
    const { container } = render(<AppStatusBadge status="stopped" />);
    const dot = container.querySelector(".animate-pulse");
    expect(dot).toBeNull();
  });
});

describe("VMStatusBadge", () => {
  it("shows Online for online status", () => {
    render(<VMStatusBadge status="online" />);
    expect(screen.getByText("Online")).toBeInTheDocument();
  });

  it("shows Unreachable for unreachable status", () => {
    render(<VMStatusBadge status="unreachable" />);
    expect(screen.getByText("Unreachable")).toBeInTheDocument();
  });

  it("shows Unknown for unknown status", () => {
    render(<VMStatusBadge status="unknown" />);
    expect(screen.getByText("Unknown")).toBeInTheDocument();
  });
});
