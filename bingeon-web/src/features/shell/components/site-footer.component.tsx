import { Link } from "@tanstack/react-router";

/** TMDB's logo and notice, as its terms ask, and the privacy page. */
export function SiteFooter() {
  return (
    <footer className="mt-16 border-t">
      <div className="mx-auto flex max-w-(--content-width) flex-col gap-3 px-(--page-gutter) py-8 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <a href="https://www.themoviedb.org" target="_blank" rel="noopener noreferrer">
            <img src="/tmdb-logo.svg" alt="TMDB" className="h-3" />
          </a>
          <p>This product uses the TMDB API but is not endorsed or certified by TMDB.</p>
        </div>
        <Link to="/privacy" className="hover:text-foreground">
          Privacy
        </Link>
      </div>
    </footer>
  );
}
