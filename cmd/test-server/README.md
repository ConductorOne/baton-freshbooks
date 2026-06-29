# Test server

Mock FreshBooks API for `baton-freshbooks`, used by CI when no real-tenant
credentials are available. Replicates the upstream API's auth flow, identity
and team-member endpoints, pagination, and error envelopes.

## Auth

| Real API | Test server |
|---|---|
| OAuth2 refresh_token grant at `https://api.freshbooks.com/auth/oauth/token` | Same flow at `http://localhost:8765/oauth/token` |
| Static bearer (access token) on every request | Hardcoded: `test-access-token` |
| `client_id` / `client_secret` / `refresh_token` from the customer | Hardcoded: `test-client-id` / `test-client-secret` / `test-refresh-token` |

Both auth modes converge on the bearer `test-access-token`, which every
resource endpoint requires.

## Endpoints

| Path | Method | Doc URL |
|---|---|---|
| `/oauth/token` | POST | https://www.freshbooks.com/api/authentication |
| `/api/v1/users/me` | GET | https://www.freshbooks.com/api/me_endpoint |
| `/api/v1/businesses/{businessID}/team_members` | GET | https://www.freshbooks.com/api/team_members |

## Seed data

Business `12345` ("Acme Test Co") with **55 team members**:

- One per known role: `owner`, `business_manager`, `business_employee`,
  `contractor`, `no_seat_employee`.
- A second `business_employee` (inactive) — a role assigned to ≥ 2 users.
- One member with **no role** (empty `business_role_name`) — the empty-grants
  path (no grant emitted).
- One member with an **unknown role** (`ghost_role`) — `newRoleResource`
  returns nil, so no grant.
- 47 filler members (`business_employee`) to push the total past the
  connector's page size (50), exercising pagination across a page boundary.

Expected sync result: **55 users**, **5 roles**, and **53 role grants**
(55 members − 1 no-role − 1 unknown-role). See `seeds.go`.

## Running locally

```bash
# Start the test server (from project root)
go run ./cmd/test-server/

# In a separate terminal, build and point the connector at it (access-token mode)
go build -o baton-freshbooks ./cmd/baton-freshbooks
./baton-freshbooks \
  --access-token test-access-token \
  --base-url http://localhost:8765 \
  --file sync.c1z

# Or exercise the OAuth refresh-token flow instead
./baton-freshbooks \
  --refresh-token test-refresh-token \
  --freshbooks-client-id test-client-id \
  --freshbooks-client-secret test-client-secret \
  --base-url http://localhost:8765 \
  --file sync.c1z

# Inspect the output
baton resources --file sync.c1z
baton grants --file sync.c1z
```

## Curl examples

```bash
# Get a token (refresh_token grant)
curl -s -X POST http://localhost:8765/oauth/token \
  -d 'grant_type=refresh_token&client_id=test-client-id&client_secret=test-client-secret&refresh_token=test-refresh-token'

# Resolve the business ID
curl -s http://localhost:8765/api/v1/users/me \
  -H 'Authorization: Bearer test-access-token'

# List team members (page 2, 50 per page)
curl -s 'http://localhost:8765/api/v1/businesses/12345/team_members?page=2&per_page=50' \
  -H 'Authorization: Bearer test-access-token'
```
