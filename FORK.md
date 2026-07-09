# Fork allocation ledger — xray-core

This fork (`kirisame-meguru/xray-core`) adds the **per-user-per-inbound traffic stats** feature directly
on **`main`** (`main` is the feature branch). The feature is switched on per inbound, by the config
itself: `$.inbounds[].trackTrafficPerUser: true`. `infra/conf` collects the tags of the opted-in
inbounds into `dispatcher.Config.TrackedInboundTags`, and the dispatcher emits `useri>>>` counters
for exactly those tags. Nothing outside the xray config controls it — the panel and node do not.

`main` is kept at **exactly one commit ahead of upstream, zero behind**: the feature is a single
commit rebased onto `xtls/main`. Never let fixes accumulate as extra commits — amend or squash into
the one feature commit, so the fork's whole delta is always one `git show` away.

To survive indefinite rebasing onto upstream, every
identifier this fork allocates from a *shared, sequential namespace* is parked in a **reserved high
band** so it can never collide with upstream's next-sequential pick. See `../FORK-RESILIENCE.md` for
the sync playbook.

## ⚠ Upstream is now `XTLS/Xray-core` (not `remnawave/xray-core`)

`remnawave/xray-core` no longer exists — Remnawave dropped its own xray fork. node 2.8.0 installs
**stock `XTLS/Xray-core`** (`Dockerfile` `UPSTREAM_REPO=XTLS`). So this feature is rebased directly
onto `XTLS/Xray-core`:

```
git remote add xtls https://github.com/XTLS/Xray-core.git   # once
git fetch xtls --tags
git rebase --onto xtls/main <old-base> main
```

Current base: **`2323273e`** (`xtls/main`, 11 commits past tag `v26.7.28`). The fork's binary is cut
as release **`v26.7.28-perinbound1`** on `kirisame-meguru/xray-core` — this tag must equal node's
`docker/Dockerfile` `ARG XRAY_CORE_VERSION`. Bump the `-perinboundN` suffix (and re-cut) whenever the
XTLS base moves or the feature commit changes.

Upstream API drift absorbed at this base: `0bafca94` ("Stats: Fix GetOrRegister*() races") replaced the
free function `stats.GetOrRegisterCounter(mgr, name)` with the method `mgr.GetOrRegisterCounter(name)`.
The feature's three `useri>>>` registration sites were moved to the method form. This class of change
**auto-merges clean and then fails to compile** — the counter calls sit far from upstream's edits, so
git sees no conflict. `go build ./...` is the guard; never skip it after a rebase.

## Why this matters

If the fork used the next-available value (e.g. proto field 4) and upstream — unaware of the fork —
later allocated field 4 to a *different* field, the two additions can auto-merge into different spots
with **no git conflict**, leaving a reused proto field number. That is a silent on-the-wire
incompatibility. Reserved-high values make that impossible.

## Allocation table

| Namespace | Symbol | File | Fork value | Upstream-conventional (PR-time) |
|-----------|--------|------|-----------|----------------------------------|
| proto field # — dispatcher `Config` | `tracked_inbound_tags` | `app/dispatcher/config.proto` | **50** | 2 |
| stats counter string ns | `useri>>>{email}>>>inbound>>>{tag}>>>traffic>>>{up\|down}link` | `app/dispatcher/default.go` | reserved prefix `useri>>>` | (unchanged) |
| JSON conf field | `trackTrafficPerUser` on `$.inbounds[]` | `infra/conf/xray.go` | (string, no band) | (unchanged) |

Field band **50–99** is declared reserved for fork features (marker comment in each `.proto`). The
counter prefix `useri>>>` is intentionally disjoint from upstream's `user>>>` / `inbound>>>` /
`outbound>>>` and is parsed with a strict 6-segment guard — no change needed, documented only.

Descriptive identifiers (JSON field names, struct field names) are unique strings: a collision would
be a *visible* git textual conflict, so they need no reserved band.

### Retired allocations — never reuse

`Policy.Stats` fields **50** (`user_inbound_uplink`) and **51** (`user_inbound_downlink`) in
`app/policy/config.proto`, together with their `statsUserInbound{Uplink,Downlink}` JSON flags, were
removed once the per-inbound `trackTrafficPerUser` flag became the sole gate for the counters. This
ledger is the only record of them: `app/policy/` is now byte-identical to upstream (a `reserved 50,
51;` marker would have kept the generated `config.pb.go` diverged forever, for a guard that only ever
protects the fork from itself). **If this fork ever needs a `Policy.Stats` field again, start at 52.**

## Generated artifacts

`app/dispatcher/config.pb.go` is generated and is the only `.pb.go` this fork touches. **Never
hand-merge it.** On any rebase conflict there, resolve the `.proto`, then regenerate (toolchain pinned
to the committed headers — protoc **v33.5**, `protoc-gen-go` **v1.36.11**, go **1.26.3**):

```
PATH="/tmp/protoc/bin:/tmp/gobin:$PATH" \
protoc --go_out=. --go_opt=paths=source_relative \
  --plugin=protoc-gen-go=/tmp/gobin/protoc-gen-go \
  app/dispatcher/config.proto
```

## Renumber-at-PR

To upstream this feature, map every **Fork value → Upstream-conventional** above (dispatcher
`tracked_inbound_tags` 50→2), remove the `[remnawave-fork]` marker comments, regenerate the `.pb.go`,
then `go build ./...`.
