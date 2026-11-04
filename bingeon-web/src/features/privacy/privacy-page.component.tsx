/** What bingeon keeps, and how to have it deleted: Google's consent screen links here. */
export function PrivacyPage() {
  return (
    <article className="max-w-prose space-y-4 leading-relaxed">
      <h1 className="type-page-title">Privacy</h1>
      <p>Browsing and searching need no account, and keep nothing about you.</p>
      <h2 className="type-section-title">If you sign in with Google</h2>
      <ul className="list-disc space-y-2 pl-5">
        <li>Your sign-in (Amazon Cognito) keeps your email address, and nothing else from Google.</li>
        <li>bingeon keeps your favourites (each movie's title, year and poster), your last AI picks, and how many of each you made today.</li>
        <li>
          To make AI picks, the titles and years of up to 100 of your favourites are sent to Google's Gemini. Nothing else about you is
          sent. Google may use what it is sent to improve its products.
        </li>
        <li>Nothing is sold or shared with anyone else, and there are no ads or trackers.</li>
      </ul>
      <h2 className="type-section-title">Deleting it</h2>
      <p>
        Delete account, in the account menu, deletes your favourites, your picks and your sign-in, with its email, at once and for
        good.
      </p>
      <p className="text-sm text-muted-foreground">
        Movie data and images come from TMDB. This product uses the TMDB API but is not endorsed or certified by TMDB.
      </p>
    </article>
  );
}
