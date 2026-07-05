# Fork allocation ledger — xray-core

This fork (`kirisame-meguru/xray-core`) adds the **per-user-per-inbound traffic stats** feature directly
on **`main`** (`main` is the feature branch). To survive indefinite rebasing onto upstream, every
identifier this fork allocates from a *shared, sequential namespace* is parked in a **reserved high
band** so it can never collide with upstream's next-sequential pick. See `../FORK-RESILIENCE.md` for
the sync playbook.

## ⚠ Upstream is now `XTLS/Xray-core` (not `remnawave/xray-core`)

`remnawave/xray-core` no longer exists — Remnawave dropped its own xray fork. node 2.8.0 installs
**stock `XTLS/Xray-core`** (`Dockerfile` `UPSTREAM_REPO=XTLS`), currently **`v26.6.27`**. So this
feature is rebased directly onto `XTLS/Xray-core`:

```
git remote add xtls https://github.com/XTLS/Xray-core.git   # once
git fetch xtls --tags
git rebase --onto <new-xtls-tag> <old-base> main
```

Current base: **`v26.6.27`**. The fork's binary is cut as release **`v26.6.27-perinbound1`** on
`kirisame-meguru/xray-core` — this tag must equal node's `Dockerfile` `ARG XRAY_CORE_VERSION`. Bump
the `-perinboundN` suffix (and re-cut) whenever the XTLS base moves.

## Why this matters

If the fork used the next-available value (e.g. proto field 4) and upstream — unaware of the fork —
later allocated field 4 to a *different* field, the two additions can auto-merge into different spots
with **no git conflict**, leaving a reused proto field number. That is a silent on-the-wire
incompatibility. Reserved-high values make that impossible.

## Allocation table

| Namespace | Symbol | File | Fork value | Upstream-conventional (PR-time) |
|-----------|--------|------|-----------|----------------------------------|
| proto field # — `Policy.Stats` | `user_inbound_uplink` | `app/policy/config.proto` | **50** | 4 |
| proto field # — `Policy.Stats` | `user_inbound_downlink` | `app/policy/config.proto` | **51** | 5 |
| proto field # — dispatcher `Config` | `tracked_inbound_tags` | `app/dispatcher/config.proto` | **50** | 2 |
| stats counter string ns | `useri>>>{email}>>>inbound>>>{tag}>>>traffic>>>{up\|down}link` | `app/dispatcher/default.go` | reserved prefix `useri>>>` | (unchanged) |

Field band **50–99** is declared reserved for fork features (marker comment in each `.proto`). The
counter prefix `useri>>>` is intentionally disjoint from upstream's `user>>>` / `inbound>>>` /
`outbound>>>` and is parsed with a strict 6-segment guard — no change needed, documented only.

Descriptive identifiers (JSON policy flags `statsUserInbound{Uplink,Downlink}`, struct field names,
the new `Dispatcher` conf section) are unique strings: a collision would be a *visible* git textual
conflict, so they need no reserved band.

## Generated artifacts

`app/policy/config.pb.go` and `app/dispatcher/config.pb.go` are generated. **Never hand-merge them.**
On any rebase conflict there, resolve the `.proto`, then regenerate (toolchain pinned to the committed
headers — protoc **v33.5**, `protoc-gen-go` **v1.36.11**, go **1.26.3**):

```
PATH="/tmp/protoc/bin:/tmp/gobin:$PATH" \
protoc --go_out=. --go_opt=paths=source_relative \
  --plugin=protoc-gen-go=/tmp/gobin/protoc-gen-go \
  app/policy/config.proto app/dispatcher/config.proto
```

## Renumber-at-PR

To upstream this feature, map every **Fork value → Upstream-conventional** above (50→4, 51→5, 50→2),
remove the `[remnawave-fork]` marker comments, regenerate the `.pb.go`, then `go build ./...`.
