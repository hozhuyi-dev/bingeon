import { canonicalQuery } from "./catalog.client";

describe("canonicalQuery", () => {
  test("sorts names and leaves out what is empty or the default", () => {
    expect(canonicalQuery({ page: 2, list: "popular", genre: undefined })).toBe("list=popular&page=2");
    expect(canonicalQuery({ page: 1, list: "popular" })).toBe("list=popular");
  });

  test("asks a search one way however it was typed", () => {
    expect(canonicalQuery({ q: "  Pulp   FICTION " })).toBe("q=pulp+fiction");
  });
});
