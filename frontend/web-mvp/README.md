# Tomo web MVP tester

This lightweight frontend lets us exercise the Go API while the full product experience comes together. Everything runs client‑side: drop the HTML file into a browser and point it at a locally running backend.

## What’s here

- Manual JWT + base URL configuration that persists in `localStorage`.
- A focus session timer that posts directly to `POST /sessions`.
- Session list + aggregate stats powered by `GET /sessions`.
- Quick health check for authenticated requests via `GET /me`.

The layout intentionally mirrors the backend routes, so we can validate behaviour without bootstrapping the existing React Native codebase.

## How to use it

1. Start the Go API (`backend/main.go`) so it serves on `http://localhost:8080` or your preferred address.
2. Generate a JWT using the existing tooling (e.g. Google sign‑in flow, manual token helper, or `utils.CreateToken` from the REPL) and paste it into the **JWT access token** field.
3. Hit **Save settings**. Tokens stick around between refreshes.
4. Use **Start Session** / **Stop & Save** to post a session. The timer produces ISO 8601 timestamps that match the backend’s expectations.
5. Select **Refresh Sessions** at any time to review persisted sessions and the `total_minutes` / `total_hours` aggregate.

Open `frontend/web-mvp/index.html` directly in the browser or serve the directory with any static file server (`npx serve frontend/web-mvp` etc.).

## Future integration work

1. **Google OAuth in the browser** — introduce the Google Identity Services script so users can complete the flow client-side, then exchange the ID token with `POST /auth/google` to receive our JWT automatically.
2. **JWT lifecycle management** — wrap API calls with helpers that read, refresh, and revoke tokens using the backend middleware once refresh endpoints land. Persist the token with stricter storage (e.g. `httpOnly` cookies through a proxy) when we graduate past this dev tool.
3. **Error surfacing** — map common backend errors (expired token, validation issues) into inline UI states instead of relying on the log panel.
4. **Session history polish** — add deletion controls (for `DELETE /sessions/{id}`) and pagination when the API exposes it.

Keep this file updated as we add polish or wire the interface into the production authentication story.
