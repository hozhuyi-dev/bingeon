/** TMDB's image CDN; a path is what the API gives ("/abc.jpg"). */
const IMAGES = "https://image.tmdb.org/t/p";

export function posterUrl(path: string, width: "w185" | "w342" | "w500" = "w342"): string {
  return `${IMAGES}/${width}${path}`;
}

export function backdropUrl(path: string): string {
  return `${IMAGES}/w1280${path}`;
}

export function profileUrl(path: string): string {
  return `${IMAGES}/w185${path}`;
}
