# bingeon

Find your next movie: browse and search the TMDB catalogue without signing in; sign in with Google to keep favourites and get AI picks from them.

<!--
  Visitor section, once live: pitch, screenshots, highlights, and the
  "Viewing only. All rights reserved." notice, as in bizzy's README. Then:
  "The rest of this README is the design, written for whoever works on it next."
-->

> [!NOTE]
> **Being built**, in [phases](#build-phases).

## Design principles

1. **Public and paid by one person, so everything has a ceiling.** Daily caps per user and overall on AI picks (inside Gemini's free tier), caps on favourites, a maximum table throughput, short Lambda timeouts. A capped request is refused with what it hit and when it frees up.
2. **Browsing never reaches Lambda twice.** The catalogue is the same for everyone, so CloudFront caches it: a page anyone opened in the last hours costs no Lambda or TMDB call. The web asks in one canonical form, and CloudFront keys only on the parameters the catalogue reads.
3. **A page is one call.** Home, a movie (via TMDB's `append_to_response`) and the user's account each come back whole. Replies are views trimmed to what the web shows.
4. **A signed-in request reads one item and writes at most one.** All a user has is one item: read it, apply the rule in Go, write it back if no one else did meanwhile. A mutation replies with the new account view, so the web reads nothing again.
5. **No key reaches the browser.** The TMDB token and Gemini key are SSM SecureStrings, read at a cold start.
6. **The user is who signed the call.** The API takes the Cognito identity of the invocation, never an id in the payload. bingeon stores no email or name, and a user can delete their account.
7. **No business rule in the web.** The account view carries `can` per action (ok, or why not and when it frees up); the web disables and explains by it.

Conventions: identifiers say "user"; British spelling (`favourites`); "today" is the Pacific-time day, when Gemini's quota resets, shown to the user in their own time.

## Tech stack

| Folder | What | Stack |
| --- | --- | --- |
| `bingeon-web` | The app (SPA) | React 19, Vite, TanStack Router / Query / Form, Tailwind, shadcn/ui, Amplify auth, Vitest |
| `bingeon-lambda` | Catalogue proxy, signed-in API | Go, DynamoDB (single table), TMDB API v3, Gemini (`google.golang.org/genai`), `pgregory.net/rapid` |
| `bingeon-infra` | AWS resources | CDK (TypeScript) |

- **Anonymous**: `GET /catalog/<action>?…` on the site's CloudFront, whose origin is the `catalog` Lambda's Function URL, which takes only CloudFront's signed requests (OAC).
- **Signed in**: Google via Cognito, open to anyone. Identity-pool credentials may invoke only the `api` Lambda, called directly from the browser, as in bizzy.
- **AI picks**: `gemini-3.5-flash-lite` (free tier: 500 requests a day, 15 a minute; `gemini-3.8-flash` allows only 20 and 5) with a JSON schema. The key's Google project has no billing, so it costs nothing; Google may use the prompts (movie titles) to improve its products. Quotas are per project, so each environment has its own.

## Code layout

```
bingeon-infra/   bin/bingeon-infra.ts → lib/{web,auth,data,api}-stack.ts, bin/bingeon-ci.ts → lib/ci-stack.ts, lib/common/, utils/
bingeon-lambda/  cmd/{catalog,api}/main.go, internal/{app,features,platform,tmdb,gemini,testutil}/
bingeon-web/     src/main.tsx, src/routes/ (file routes), src/features/<feature>/
scripts/         changed.mjs (what a PR or deploy touches), web-env.mjs, secrets.mjs
```

### bingeon-infra

[bin/bingeon-infra.ts](bingeon-infra/bin/bingeon-infra.ts) reads `.env.<ENVIRONMENT>` and creates `bingeon-<env>-{web,auth,data,api}`:

- **web**: what anyone reaches: the site's bucket, CloudFront (`/catalog/*` to the `catalog` Lambda, also here so the stack needs nothing later), whose URL is Cognito's first callback. Development's bucket stays empty; `vite dev` proxies `/catalog` to it.
- **auth**: Cognito user pool (Google only, no self sign-up) and identity pool, one role for every signed-in user, no guests.
- **data**: the table, on demand with a maximum throughput, TTL on `expires`.
- **api**: the `api` Lambda and the role's permission to invoke it (owned here, so the dependency stays one-way).

`catalog` can read its SSM parameter only, never the table. No Lambda reserves concurrency: the account's limit is 10, and AWS keeps 10 unreserved.

[bin/bingeon-ci.ts](bingeon-infra/bin/bingeon-ci.ts) (`bingeon-ci`, deployed by hand with `pnpm ci:deploy`) holds `bingeon-production-github`, the role [deploy.yml](.github/workflows/deploy.yml) assumes through the account's GitHub OIDC provider (shared, so only looked up). Only the `production` GitHub environment can assume it, and it may do only what the workflow does outside CloudFormation.

### bingeon-lambda

Both commands use bizzy's router (`{action, payload}` to a `Query` or `Mutation`, `{data}` or `{error}` back, `apperr` codes), wired by `internal/app`.

- **`cmd/catalog`** maps `GET /catalog/<action>?<query>` to `{action, payload}`. Each action is registered with its TTL (`router.Cached`), replied as `public, max-age=300, s-maxage=<ttl>`; `invalid` and `not_found` are kept a minute, a TMDB failure `no-store`.
- **`cmd/api`** takes the user from the invocation's Cognito identity (`lambdacontext`) and refuses a call without one.
- Each logs one `request` line: action, ms, table reads and writes, TMDB and Gemini calls.

**Features** (`internal/features/<feature>/`: `api.go`, the entity file, `actions.go` with each `Can…`):

| Feature | Actions | Notes |
| --- | --- | --- |
| `catalog` | `catalog.home`, `movies.list`, `movies.search`, `movies.get` | Views over `internal/tmdb`; no table |
| `account` | `account.get`, `account.delete` | Owns the `Account` item, its view and every limit (`account/limits.go`) |
| `favourites` | `favourites.add`, `favourites.remove` | Takes a movie id; the API reads its title, year and poster from TMDB |
| `picks` | `picks.make` | Gemini over the favourites, each title found in TMDB |

`favourites` and `picks` change the account only through `account.Change(ctx, id, fn)`: read, apply, write if unchanged, else again. `app/dependencies_test.go` enforces imports.

**Storage**: one table, no index.

| PK | SK | Holds |
| --- | --- | --- |
| `USER#<identityId>` | `ACCOUNT` | Favourites (id, title, year, poster, added), the last picks with reasons, today's counts, `version` |
| `QUOTA#<yyyy-mm-dd>` | `GEMINI` | The day's Gemini calls; `expires` two days on |

`platform/table` is smaller than bizzy's: `Get` (eventually consistent; a stale read only makes the write retry) and `Commit`, all or none, of `Put` or `Delete` at a version and `Count` under a cap. `table.Memory` and DynamoDB pass one contract test.

**`picks.make`**: check `can.pick`; reserve the user's and the global count in one commit; ask Gemini for 6 movies like the 100 newest favourites (TMDB's titles, never user text), excluding them and the last picks, each with a reason; find each in TMDB in parallel; keep them as the last picks. On a Gemini or TMDB failure the user's count is given back; a 429 is "busy, try again in a minute".

**Tests**: `testutil/testapp` is the whole app on the memory table with fake TMDB and Gemini, called with JSON as the web calls it. A small `rapid` test runs several users at random and checks no cap is passed, a refusal changes nothing, and `can` matches the action.

Stored types change additively only; the first change that can't brings a schema version and an upgrade on read.

### bingeon-web

`src/main.tsx` → `router.tsx` → `src/routes/` (file routes that only read params and render a page). bizzy's conventions:

- `src/features/<feature>/` files are `<kebab-name>.<layer>.ts(x)`; layers in order: `types`, `constant`, `config`, `util`, `form`, `service`, `query` / `mutation`, `hook`, `component`. Pages are `x-page.component.tsx` at the feature root; components reach data only through `query`, `mutation` and `hook`. `types` mirror `api.go` (`AccountView`).
- `core/` is infrastructure, `components/shared/` shared components by role, `components/ui/` shadcn's, untouched. Neither imports a feature (ESLint).
- `core/` has two clients: `catalog.client.ts` (canonical query: keys sorted, search words trimmed, lowercased, single-spaced) and `lambda.client.ts`, both throwing `ApiError`.
- The catalogue's `staleTime` matches its `s-maxage`; the account is read once a session and replaced by each mutation's reply.
- The URL holds the state (search, list, genre, page); no Zustand or Zod.
- A heart tapped signed out signs in, then favourites the movie (history state through the redirect).
- Loading, empty and error states are the pigeon's drawings; a page's data loads in its route's loader, so nothing jumps.
- Tokens only in `styles/app.css`. Posters from TMDB's image CDN. The footer has TMDB's logo and notice, as its terms ask.

### Left out from bizzy

- **`txn`**: every commit goes through one versions item, a hot spot with many users; no rule here spans two items.
- **`locked`, the simulation**: no edit forms, few rules; a small `rapid` test instead.
- **Migrations, backups, data copy and replay**: favourites can be made again, and types change additively.
- **`changed`/`versions`**: a mutation returns the whole account.
- **PWA**: a site people visit, not install.
- **PDFs, S3 files, Google Calendar, jobs**: no such features.

## Features and limits

- **Anyone**: Home (Popular, Now playing and Upcoming in Malaysia, `region=MY`, Top rated), lists by genre, search, a movie's page with cast, trailer, similar and recommended.
- **Signed in**: favourites (a heart on every card), AI picks from them, delete my account.

Expected users: 10 at most. Limits are constants with their reasons, in `account/limits.go` or the stacks.

| Limit | Value | Why |
| --- | --- | --- |
| AI picks per user per day | 20 (6 movies each) | 10 users at the cap use half the global cap |
| Gemini calls per day, all users | 400 | 80% of the free 500, so Google never refuses for the day |
| Favourites per user | 500 | Keeps the item near 50 KB (DynamoDB allows 400) |
| Favourite changes per user per day | 1,000 | Stops a script |
| Favourites sent to Gemini | 100 newest | About 4,000 tokens a call |
| Lambda timeouts | `catalog` 5 s, `api` 20 s | Picks wait on Gemini |
| Table throughput | 10 reads, 10 writes a second | Caps what a bug or flood bills |
| Search | 2–100 characters, page ≤ 500 | TMDB's page limit; bounded cache keys |

Cache (`s-maxage`): home and lists 6 hours, search and a movie 24 hours; browsers 5 minutes.

## Environments

| Environment | Deployed by | Web |
| --- | --- | --- |
| `development` | `pnpm dev:deploy` from a laptop | `pnpm web:dev`, http://localhost:4200 |
| `production` | GitHub Actions, on push to `main` | its CloudFront URL |

`main` is production; pull requests are checked by CI. AWS profile `hozhuyi.dev`, `ap-southeast-1`; GitHub `hozhuyi-dev/bingeon`. Cost alerts are account-wide, not here. Keys are never in GitHub: `pnpm secrets:set <env>` puts them in SSM at `/bingeon/<env>/{tmdb-token,gemini-key}`. A web deploy invalidates only unhashed files, never `/catalog/*`.

## Development setup

Node 24+, pnpm 10, Go (see `go.mod`), the AWS CLI. Then `pnpm install`. The rest (Google OAuth client for sign-in, TMDB token, Gemini key, `.env.development`) comes with the infra phase.

## Commands

```sh
pnpm web:dev             # http://localhost:4200
pnpm web:test | web:lint | web:typecheck
pnpm infra:test | infra:lint | infra:typecheck
pnpm lambda:test | lambda:vet
pnpm cdk:diff | cdk:deploy | cdk:gc
```

## If the catalogue is flooded

An uncached request (a new search) is a Lambda and a TMDB call, capped only by the account's 10 concurrent executions, so **a flood throttles bizzy and traceform too**. Accepted: the cache absorbs normal use and TMDB's rate limit slows a flood. If the account limit is raised, reserve 5 for `catalog` and 2 for `api`. To stop it at once:

```sh
aws lambda put-function-concurrency --function-name bingeon-<env>-catalog \
  --reserved-concurrent-executions 0 --profile hozhuyi.dev --region ap-southeast-1
aws lambda delete-function-concurrency --function-name bingeon-<env>-catalog \
  --profile hozhuyi.dev --region ap-southeast-1   # back on
```

## Build phases

1. ~~Scaffold and tooling~~: done.
2. **Infra and auth**: the four stacks, `bingeon-ci`, sign-in in development (confirming the API sees the Cognito identity).
3. **API**: router, table, `tmdb`, `gemini`, the features and limits, with tests.
4. **Web**: catalogue, movie page, search, favourites, picks, account.
5. **CI/CD**: workflows, the `production` environment, the first deploy (when asked).

## Open questions

- **CloudFront's free flat-rate plan** for production (WAF rate limiting, no overage), taken by hand in the console after the first deploy, if offered and if it includes rate limiting. No distribution in the account had a web ACL on 2 Nov 2026, so the three plan slots look free.
