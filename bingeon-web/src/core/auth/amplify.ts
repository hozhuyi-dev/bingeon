/**
 * Amplify (sign-in) loads only when it is needed: someone signs in, comes
 * back from signing in, or was signed in before. Browsing without signing
 * in never downloads it.
 */

import type * as AmplifyAuth from "aws-amplify/auth";

type Auth = typeof AmplifyAuth;

let loading: Promise<Auth> | undefined;

/** Amplify's auth functions, Amplify configured. */
export function loadAuth(): Promise<Auth> {
  loading ??= (async () => {
    const [{ Amplify }, auth] = await Promise.all([
      import("aws-amplify"),
      import("aws-amplify/auth"),
      // Completes a sign-in when the page is Cognito's redirect back.
      import("aws-amplify/auth/enable-oauth-listener"),
    ]);
    // Redirects follow whichever origin the app is served from; each must
    // be in the user pool's callback and logout URLs.
    const { origin } = window.location;
    Amplify.configure({
      Auth: {
        Cognito: {
          userPoolId: import.meta.env.VITE_AMPLIFY_USER_POOL_ID,
          userPoolClientId: import.meta.env.VITE_AMPLIFY_USER_POOL_CLIENT_ID,
          identityPoolId: import.meta.env.VITE_AMPLIFY_IDENTITY_POOL_ID,
          loginWith: {
            oauth: {
              domain: import.meta.env.VITE_AMPLIFY_OAUTH_DOMAIN,
              redirectSignIn: [`${origin}/`],
              redirectSignOut: [`${origin}/`],
              responseType: "code",
              // The last lets a user delete their own sign-in.
              scopes: ["email", "openid", "aws.cognito.signin.user.admin"],
            },
          },
        },
      },
    });
    return auth;
  })();
  return loading;
}

/**
 * Whether there may be a sign-in to resume, without loading Amplify: its
 * tokens are in storage, or this page is Cognito's redirect back.
 */
export function mayBeSignedIn(): boolean {
  if (new URLSearchParams(window.location.search).has("code")) return true;
  const prefix = `CognitoIdentityServiceProvider.${import.meta.env.VITE_AMPLIFY_USER_POOL_CLIENT_ID}.`;
  try {
    return Object.keys(localStorage).some((key) => key.startsWith(prefix));
  } catch {
    return false;
  }
}
