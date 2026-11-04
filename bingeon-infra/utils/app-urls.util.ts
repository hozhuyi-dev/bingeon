/**
 * Where the app runs, as Cognito's callback and logout lists want them.
 *
 * `siteOrigin` is the site ("https://d123.cloudfront.net", a deploy-time
 * token), first so it is Cognito's default redirect. `extra` are the URLs
 * from the env file, each ending in "/" (development's localhost).
 */
export function appUrls(siteOrigin: string, extra: { callbackUrls: Array<string>; logoutUrls: Array<string> }) {
  const site = `${siteOrigin}/`;
  return {
    callbackUrls: [site, ...extra.callbackUrls],
    logoutUrls: [site, ...extra.logoutUrls],
  };
}
