# Issue #17 — Go/Wails native notification integration plan
Date: 2026-11-04
Reference: https://github.com/hugoalves-dev/OrganizerV2/issues/17

## Scope and design
Keep OrganizerV2 watcher/CLI Go 1.25, headless and CGO-disabled. Build the Wails v3 beta.32 UI in its own Go 1.27.2/CGO module at `cmd/organizer-popup`, with `eaxum/wails-dragout` pinned to the upstream reviewed revision. Reuse the notifier's existing event registry, action handler and collision-safe file service through a short-lived loopback bridge. Native notifications remain the fallback.

## Phase 1 — approved PoC
1. **Investigation/design.** Inspect `internal/notifier`, existing actions/callbacks, `go.mod`, CI, Wails/Wails-dragout licensing and runtime. Compare options. **Done in preceding PoC**, documented in project conversation.
2. **Isolated prototype.** Native copy-only drag-out with Go, Wails v3 and `wails-dragout`; do not alter the watcher. **Done**, user subsequently supplied a corrected PoC archive in issue comment #6646880909.
3. **Validation/report.** Windows 11 manual copy to Explorer reportedly succeeded (comment #6646734242). Unicode, cancellation, Linux X11/Wayland and detailed native effect negotiation remain separate verification items. **Approval checkpoint:** user explicitly approved tasks 4–6 in issue comment #6646880909 and in the ChatGPT request.

## Phase 2 — authorized implementation
4. **Custom notifier.** Add popup launch with readiness handshake, bounded concurrent windows, one-time event IDs, authenticate callback to shared action handlers, frameless borderless Wails window and copy-only drag. Tests: helper source validation, Windows fallback and action ID validation.
5. **Config and fallback.** Default `notifications.provider` to custom for legacy YAML, support native; preserve all existing flags and shortcuts; no notifications when disabled; Linux notify-send fallback. Tests: defaults, invalid values, legacy YAML, roundtrip. Update `config init`, config example and README.
6. **Build/legal/hardening.** Keep CGO optional; add Windows helper build and license-bearing release bundle; preserve upstream MIT/AGPL text and notices, document acceptance/troubleshooting; run Windows/Linux CI and manually exercise Windows 11, DPI, Explorer/Desktop, file sizes, Unicode, cancellation and shortcut collisions.

## Change to product scope from issue comments
Preferred default Move, Ctrl Copy, Alt Link and Shift Move require an OLE effect-negotiation backend not present in upstream `wails-dragout` (copy-only). Do not silently pretend `dropped` means a successful move or delete the source. The issue author permitted continuing if Move is unavailable; this delivery retains safe copy-only until a native backend and its test matrix can be verified.

## Remaining verification
Automated CI results and artifact builds are recorded on PR #18. Runtime testing on physical Windows 11, alternative DPI/taskbar placement, and Linux X11/Wayland are not available from the current environment. Do not claim those manual checks passed. Review AGPL obligations before distribution.
