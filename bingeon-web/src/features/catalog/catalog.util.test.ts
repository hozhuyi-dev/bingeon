import { parseBrowseSearch, parseSearchSearch } from "./catalog.util";

describe("search params", () => {
  test("browse is a genre, a known list, or popular", () => {
    expect(parseBrowseSearch({ genre: "18", page: "3" })).toEqual({ genre: 18, page: 3 });
    expect(parseBrowseSearch({ list: "upcoming" })).toEqual({ list: "upcoming", page: undefined });
    expect(parseBrowseSearch({ list: "trending", page: "1" })).toEqual({ list: "popular", page: undefined });
  });

  test("pages outside 2 to 500 are the first", () => {
    expect(parseSearchSearch({ q: "heat", page: "501" })).toEqual({ q: "heat", page: undefined });
    expect(parseSearchSearch({ q: "  " })).toEqual({ q: undefined, page: undefined });
  });
});
