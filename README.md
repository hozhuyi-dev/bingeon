# bingeon

Find your next movie: browse and search the TMDB catalogue without signing in; sign in with Google to keep favourites and get AI picks from them.

<!--
  Visitor section, once live: pitch, screenshots, highlights, and the
  "Viewing only. All rights reserved." notice. Then:
  "The rest of this README is the design, written for whoever works on it next."
-->

> [!NOTE]
> **Being built**, in [phases](#build-phases).

## Design principles

1. **Public and paid by one person, so everything has a ceiling.** Daily caps per user and overall on AI picks (inside Gemini's free tier), caps on favourites, a table inside DynamoDB's free tier, short Lambda timeouts. A capped request is refused with what it hit and when it frees up.
2. **Browsing never reaches Lambda twice.** The catalogue is the same for everyone, so CloudFront caches it: a page anyone opened in the last hours costs no Lambda or TMDB call. The web asks in one canonical form, and CloudFront keys only on the parameters the catalogue reads.
3. **A page is one call.** Home, a movie (via TMDB's `append_to_response`) and the user's account each come back whole. Replies are views trimmed to what the web shows.
4. **A signed-in request touches only the user's own small items.** The table has 5 write units a second, so a user is a few small items plus one per favourite: a heart writes about 4 units, never a whole list. Read, apply the rule in Go, write if no one else did meanwhile. A mutation replies with what it changed, so the web reads nothing again.
5. **No key reaches the browser.** The TMDB token and Gemini key are SSM SecureStrings, read at a cold start.
6. **The user is who signed the call.** The API takes the Cognito identity of the invocation, never an id in the payload. bingeon stores no email or name, and a user can delete their account.
7. **No business rule in the web.** The account view carries `can` per action (ok, or why not and when it frees up); the web disables and explains by it.

Conventions: identifiers say "user"; British spelling (`favourites`); "today" is the Pacific-time day, when Gemini's quota resets, shown to the user in their own time.

## Tech stack

| Folder | What | Stack |
| --- | --- | --- |
| `bingeon-web` | The app (SPA) | React 19, Vite, TanStack Router / Query / Form, Tailwind, shadcn/ui, Amplify auth, Vitest |
| `bingeon-lambda` | Catalogue proxy, signed-in API | Go, DynamoDB (single table), TMDB API v3, Gemini (`google.golang.org/genai`) |
| `bingeon-infra` | AWS resources | CDK (TypeScript) |

- **Anonymous**: `GET /catalog/<action>?…` on the site's CloudFront, whose origin is the `catalog` Lambda's Function URL, which takes only CloudFront's signed requests (OAC).
- **Signed in**: Google via Cognito, open to anyone. Identity-pool credentials may invoke only the `api` Lambda, called directly from the browser.
- **AI picks**: `gemini-3.5-flash-lite` (free tier: 500 requests a day, 15 a minute; `gemini-3.8-flash` allows only 20 and 5) with a JSON schema. The key's Google project has no billing, so it costs nothing; Google may use the prompts (movie titles) to improve its products. One Google project, `bingeon`, serves both environments: its quota is shared, so their daily caps add up to less than 500.

## Code layout

```
bingeon-infra/   bin/bingeon-infra.ts → lib/{web,auth,data,api}-stack.ts, bin/bingeon-ci.ts → lib/ci-stack.ts, lib/common/, utils/
bingeon-lambda/  cmd/{catalog,api}/main.go, internal/{app,catalog,favourites,picks,account,platform}/
bingeon-web/     src/main.tsx, src/routes/ (file routes), src/features/<feature>/
scripts/         changed.mjs (what a PR or deploy touches), web-env.mjs, secrets.mjs, aws.mjs (shared)
```

### bingeon-infra

[bin/bingeon-infra.ts](bingeon-infra/bin/bingeon-infra.ts) reads `.env.<ENVIRONMENT>` and creates `bingeon-<env>-{web,auth,data,api}`:

- **web**: what anyone reaches: the site's bucket, CloudFront (`/catalog/*` to the `catalog` Lambda, also here so the stack needs nothing later), whose URL is Cognito's first callback. Development's bucket stays empty; `vite dev` proxies `/catalog` to it.
- **auth**: Cognito user pool (Google only, no self sign-up) and identity pool, one role for every signed-in user, no guests.
- **data**: the table, provisioned at 5 read and 5 write units (`CAPACITY_UNITS`): the free tier's 25 less bizzy's 15, split between the two environments. TTL on `expires`.
- **api**: the `api` Lambda and the role's permission to invoke it (owned here, so the dependency stays one-way).

CloudFront's cache key, and all `catalog` receives, is the query parameters in `CATALOG_QUERY` ([site.ts](bingeon-infra/lib/common/web/site.ts)); one the catalogue starts reading must be added there. `catalog` can read the TMDB token only, never the table; `api` gets, queries, puts, updates and deletes items (never a scan) and reads both keys. Cognito asks Google only for `openid email` and keeps only the email. No Lambda reserves concurrency: the account's limit is 10, and AWS keeps 10 unreserved.

[bin/bingeon-ci.ts](bingeon-infra/bin/bingeon-ci.ts) (`bingeon-ci`, deployed by hand with `pnpm ci:deploy`) holds `bingeon-production-github`, the role [deploy.yml](.github/workflows/deploy.yml) assumes through the account's GitHub OIDC provider (shared, so only looked up). Only the `production` GitHub environment can assume it, and it may do only what the workflow does outside CloudFormation.

### bingeon-lambda

Vertical slices: one package per feature, each in four layers, dependencies pointing inward.

| File | Layer | Does |
| --- | --- | --- |
| `handler.go` | Transport | Registers the feature's actions; decodes a payload, calls the service, returns its view |
| `service.go` | Use cases | One method per action: reads, applies the domain's rules, writes, calls TMDB or Gemini, through small interfaces it declares (`Store`, `Movies`, `Picker`) |
| `store.go` | Repository | The feature's own DynamoDB items, implementing its `Store` |
| `domain.go` | Core | Types, limits and rules (`Can…` verdicts, the day's counts); no I/O |

| Feature | Actions | Owns |
| --- | --- | --- |
| `catalog` | `catalog.home`, `movies.list`, `movies.search`, `movies.get` | Nothing: views over TMDB |
| `favourites` | `favourites.add`, `favourites.remove` | `FAV#<movieId>` items (title, year and poster read from TMDB) and `FAVOURITES` |
| `picks` | `picks.make` | `PICKS` and the global `QUOTA#<day>` |
| `account` | `account.get`, `account.delete` | Nothing: composes `favourites` and `picks` (the only slice importing others) |

`platform/` is what slices share: `router` (`{action, payload}` in, `{data}` or `{error}` out), `apperr` (codes, `Verdict`), `dynamo`, `tmdb`, `gemini`, `ssm`, `user` (the signed-in user in the context), `day` (the Pacific day), `meter` (counts for the log). `internal/app` wires the slices, joining `picks` to `favourites` through an interface; `app/dependencies_test.go` enforces imports, and per file that `domain.go` does no I/O and only `store.go` touches DynamoDB.

- **`cmd/catalog`** maps `GET /catalog/<action>?<query>` to `{action, payload}`. Each action is registered with its TTL (`router.Cached`), replied as `public, max-age=300, s-maxage=<ttl>`; `invalid` and `not_found` are kept a minute, a TMDB failure `no-store`.
- **`cmd/api`** takes the user from the invocation's Cognito identity (`lambdacontext`), which Lambda sets from the identity-pool credentials that signed the call (confirmed in development), and refuses a call without one.
- Each logs one `request` line: action, ms, table read and write units, TMDB and Gemini calls.

**Storage**: one table, no index; each item is plain attributes.

| PK | SK | Owner | Holds |
| --- | --- | --- | --- |
| `USER#<identityId>` | `FAVOURITES` | favourites | Count, the day and its changes, `version` |
| `USER#<identityId>` | `FAV#<movieId>` | favourites | Title, year, poster, added (~150 B) |
| `USER#<identityId>` | `PICKS` | picks | The day and its picks made, the last picks with reasons, `version` (~2 KB) |
| `QUOTA#<yyyy-mm-dd>` | `GEMINI` | picks | The day's Gemini calls; `expires` two days on |

A change reads its item consistently, applies the domain's rule, and writes it back on condition its `version` hasn't moved, else tries again (three times). A heart is one transaction (the favourite and `FAVOURITES`, about 4 write units). `account.get` queries the user's `FAV#` items (500 ≈ 75 KB ≈ 10 read units, eventually consistent, once a session) and gets the three small items. A mutation replies with what it changed and the account's `status` (`can`, today's counts), so the web reads nothing again.

**`picks.make`**: check `can.pick`; reserve the user's and the global count in one transaction; ask Gemini for 8 movies like the 100 newest favourites (TMDB's titles, never user text), excluding them and the last picks, each with a reason; find each in TMDB in parallel; keep the first 6 found that aren't favourites as the last picks. On a Gemini or TMDB failure, or none found, the user's count is given back; Gemini's 429 is "busy, try again in a minute".

**Tests**: `domain.go` by plain unit tests; each service with an in-memory store and fake TMDB and Gemini; each `store.go` against a real table when `TABLE_NAME` is set (development's: its own random test user, deleted after); the catalogue's HTTP mapping through the router.

Stored types change additively only; the first change that can't brings a schema version and an upgrade on read.

### bingeon-web

`src/main.tsx` → `router.tsx` → `src/routes/` (file routes that only read params, load their page's data and render it).

| Route | Page |
| --- | --- |
| `/` | The genres as chips at the top, then four shelves (Popular, Now playing and Upcoming in Malaysia, Top rated), each with "See all" |
| `/browse?list=…` or `?genre=…`, `&page=` | The genre chips (the one showing highlighted), a grid, Previous and Next |
| `/search?q=…&page=` | The same; searches on Enter, not as you type (each new search is an uncached call) |
| `/movie/$id` | Details, cast, a "Watch trailer" link to YouTube (no embed), recommended and similar |
| `/me` | My movies: AI picks ("Make picks", what's left today, why not), then favourites; sign in when signed out |
| `/privacy` | What is kept, and deleting it; Google's consent screen links here |

- `src/features/<feature>/` files are `<kebab-name>.<layer>.ts(x)`; layers in order: `types`, `constant`, `config`, `util`, `form`, `service`, `query` / `mutation`, `hook`, `component`. Pages are `x-page.component.tsx` at the feature root; components reach data only through `query`, `mutation` and `hook`. `types` mirror the slices' views.
- `core/` is infrastructure, `components/shared/` shared components by role (the pigeon, messages, movie card, grid, shelf), `components/ui/` shadcn's, untouched. Neither imports a feature (ESLint).
- `core/` has two clients, both throwing `ApiError`: `catalog/catalog.client.ts` (canonical query: names sorted, page 1 and empty values left out, search trimmed, single-spaced, lowercased) and `lambda/lambda.client.ts`.
- **Sign-in loads only when needed** (`core/auth/amplify.ts`): Amplify and the AWS SDK are imported when someone signs in, or has tokens stored or is back from Cognito; anonymous browsing never downloads them.
- The catalogue's `staleTime` matches its `s-maxage`; the account is read once a session and patched by each mutation's reply (`account.util.ts`).
- The URL holds the state (search, list, genre, page); no Zustand or Zod.
- A heart tapped signed out signs in, then returns to the page and favourites the movie (kept in session storage through the redirect). A refused heart says why (`can`) without calling the API.
- Rows that scroll sideways (shelves, genre chips on a phone) hide the scrollbar (`scrollbar-none`): swiped on a phone, a shelf moved a screenful at a time by its ‹ › on a wider screen.
- Loading is a skeleton grid; empty, error and not-found are the pigeon with a line of text. A page's data loads in its route's loader, so nothing jumps.
- The theme follows the device's light or dark setting (`index.html`); no toggle.
- Tokens only in `styles/app.css`. Posters from TMDB's image CDN. The footer has TMDB's logo (`public/tmdb-logo.svg`) and notice, as its terms ask. The favicon is the pigeon (`public/favicon.svg`).

## Features and limits

- **Anyone**: Home (Popular, Now playing and Upcoming in Malaysia, `region=MY`, Top rated), lists by genre, search, a movie's page with cast, trailer, similar and recommended.
- **Signed in**: favourites (a heart on every card), AI picks from them, delete my account (everything kept, then the Cognito user, with its email; the client allows `aws.cognito.signin.user.admin` for that alone).

Expected users: 10 at most. Limits are constants with their reasons, in each slice's `domain.go` or the stacks.

| Limit | Value | Why |
| --- | --- | --- |
| AI picks per user per day | 20 (6 movies each) | 10 users at the cap use half the global cap |
| Gemini calls per day, all users | production 400, development 50 (`GEMINI_DAILY_LIMIT` in the api stack) | Both share the free 500, so Google never refuses for the day |
| Favourites per user | 500 | Read in about 10 units, one query page |
| Favourite changes per user per day | 1,000 | Stops a script |
| Favourites sent to Gemini | 100 newest | About 4,000 tokens a call |
| Lambda timeouts | `catalog` 5 s, `api` 20 s | Picks wait on Gemini |
| Table capacity | 5 read, 5 write units a second, each environment | DynamoDB's free tier; beyond it requests are throttled, so it also caps the bill |
| Search | 2–100 characters, page ≤ 500 | TMDB's page limit; bounded cache keys |

Cache (`s-maxage`): home and lists 6 hours, search and a movie 24 hours; browsers 5 minutes.

## Environments

| Environment | Deployed by | Web |
| --- | --- | --- |
| `development` | `pnpm dev:deploy` from a laptop | `pnpm web:dev`, http://localhost:4200 |
| `production` | GitHub Actions, on push to `main` | its CloudFront URL |

`main` is production; pull requests are checked by CI. AWS profile `hozhuyi.dev`, `ap-southeast-1`; GitHub `hozhuyi-dev/bingeon`. Cost alerts are account-wide, not here. Keys are never in GitHub or a build: `pnpm secrets:set <env>` puts them in SSM at `/bingeon/<env>/{tmdb-token,gemini-key}`, so they aren't in a template, a Lambda's configuration or a workflow's reach, and both environments get them the same way. A web deploy invalidates only unhashed files, never `/catalog/*`.

## Development setup

Node 24+, pnpm 10, Go (see `go.mod`), the AWS CLI. Then `pnpm install`, and once per environment:

1. **Google OAuth client** `bingeon-<env>`, Web application, in the Google Cloud project `bingeon` (no billing; its consent screen is set up once and asks only for email). It stays in "Testing", where only its test users can sign in, until production: publishing needs a homepage and privacy policy URL, the production site and its `/privacy` page. Authorised redirect URI: `https://bingeon-<env>.auth.ap-southeast-1.amazoncognito.com/oauth2/idpresponse`.
2. **`bingeon-infra/.env.<env>`** from [.env.example](bingeon-infra/.env.example), with that client.
3. **Keys**: a TMDB API read access token (themoviedb.org → Settings → API) and a Gemini API key from the `bingeon` project (aistudio.google.com); one of each may serve both environments. Then `pnpm secrets:set <env>`.
4. **Development**: `pnpm dev:deploy` (all stacks, then `web:env` writes `bingeon-web/.env.development`), then `pnpm web:dev`.

## Commands

```sh
pnpm web:dev             # http://localhost:4200
pnpm web:test | web:lint | web:typecheck
pnpm infra:test | infra:lint | infra:typecheck
pnpm lambda:test | lambda:vet
pnpm lambda:deploy       # development's two functions, hotswapped
pnpm dev:deploy          # development: every stack, then web:env
pnpm web:env             # bingeon-web/.env.<env> from the stacks' outputs
pnpm secrets:set <env>   # the TMDB token and Gemini key, into SSM
pnpm cdk:diff | cdk:deploy | cdk:gc
pnpm ci:diff | ci:deploy # bingeon-ci, by hand
```

Scripts act on `ENVIRONMENT` (default `development`) with profile `hozhuyi.dev`.

## If the catalogue is flooded

An uncached request (a new search) is a Lambda and a TMDB call, capped only by the account's 10 concurrent executions, so **a flood throttles the account's other functions too**. Accepted: the cache absorbs normal use and TMDB's rate limit slows a flood. If the account limit is raised, reserve 5 for `catalog` and 2 for `api`. To stop it at once:

```sh
aws lambda put-function-concurrency --function-name bingeon-<env>-catalog \
  --reserved-concurrent-executions 0 --profile hozhuyi.dev --region ap-southeast-1
aws lambda delete-function-concurrency --function-name bingeon-<env>-catalog \
  --profile hozhuyi.dev --region ap-southeast-1   # back on
```

## Build phases

1. ~~Scaffold and tooling~~: done.
2. ~~Infra and auth~~: done; development deployed. Until phase 3 the functions are stubs: `api` answers with the caller's identity, `catalog` with a 404, and the header shows the sign-in check. `bingeon-ci` is built, deployed with the CI phase.
3. ~~API~~: done; development deployed. Stores checked against the development table, the catalogue through CloudFront (cache hits), and a real Gemini pick (6 found, ~6 s).
4. ~~Web~~: done; development deployed and tried by hand.
5. **CI/CD**: workflows, the `production` environment, the first deploy (when asked), then publishing the Google consent screen with the site's URL and `/privacy`.

## Open questions

- **CloudFront's free flat-rate plan** for production (WAF rate limiting, no overage), taken by hand in the console after the first deploy, if offered and if it includes rate limiting. No distribution in the account had a web ACL on 2 Nov 2026, so the three plan slots look free.
