import "@testing-library/jest-dom/vitest";
import { vi } from "vitest";

// jsdom has no matchMedia; the theme reads the device's light/dark setting.
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  }),
});
