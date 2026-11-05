# OrganizerV2 issue #17 — production integration

## Current behavior

- `notifications.provider: custom` (default) launches `organizer-popup.exe` next to the headless Windows `organizer.exe`. It uses Wails v3 and `eaxum/wails-dragout` for a **real native copy-only drag**.
- `notifications.provider: native` uses the existing Windows toast unchanged.
- If the helper is missing, cannot start, does not send READY within 10s, or four popups are already active, the organizer logs an error and delivers a native toast. The watcher never waits on the UI thread.
- Existing Open File, Open Location and Confirm actions are routed to the **same** in-memory registry, selector resolver and file-action handler as native toasts, over an ephemeral authenticated 127.0.0.1 endpoint. The UI sees opaque destination IDs, never arbitrary target paths. The registry Claims each event exactly once.
- The popup is frameless, topmost and hidden from the Windows taskbar, positioned within the primary monitor's work area (which excludes the taskbar). Its duration, button visibility and drag availability come from YAML.
- Linux with a graphical session attempts the optional `organizer-popup` GTK3 helper when `custom` is selected and the executable is installed; otherwise it falls back to the original `notify-send` notifier. X11/Wayland drag-out remains **experimental and unverified**. `native` selects `notify-send` unconditionally.

## Build the optional helper (Windows 11)

```powershell
cd cmd/organizer-popup
# Requires Go 1.27.2+, MinGW GCC, WebView2, and Wails 3 beta.32.
go mod tidy
go run github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.32 generate bindings -d ./assets/bindings
go build -trimpath -tags production -ldflags="-H windowsgui" -o ../../dist/organizer-popup.exe .
cd ../..
# Keep organizer.exe and organizer-popup.exe side-by-side.
```

The main `go.mod` remains Go 1.25, with `CGO_ENABLED=0`; the popup lives in its own module requiring CGO and Go 1.27.2. Builds that package only the watcher may be distributed without the popup, with native notification fallback.

## Linux helper (amd64, experimental)

The upstream native drag-out backend uses **GTK3**, while modern Wails v3 defaults to GTK4. Building without the `gtk3` tag can pass an incompatible GTK4 pointer to a GTK3 drag function. Install build dependencies `pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev` on supported Debian/Ubuntu releases; runtime dependencies also apply.

```bash
cd cmd/organizer-popup
go mod tidy
go run -tags gtk3 github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.32 generate bindings -d ./assets/bindings
CGO_ENABLED=1 go test -tags gtk3 ./...
CGO_ENABLED=1 go build -trimpath -tags 'gtk3 production' -o ../../organizer-popup .
```

Keep the executable `organizer-popup` next to the Linux watcher and make it executable. Only Linux amd64 packages currently include this experimental helper. An active `DISPLAY` or `WAYLAND_DISPLAY` is required. Linux native sessions must be validated independently; a successful build or XWayland test does not establish native Wayland drag-out (the upstream implementation has input-event/serial limitations). Failure to start the helper preserves `notify-send`.

See [cross-platform manual acceptance checks](issue-17-acceptance.md).

## Safe drag-out semantics and known limitation

The current upstream `wails-dragout` is **copy-only**. A `dropped` result reports that the drop target accepted a drag, **not** that a file was copied successfully and not that a move occurred. The OrganizerV2 source path is never deleted or updated based on this return value. No Move/Copy/Link modifier claim is made.

The requirement added in issue comment #6646734242 (default Move, Ctrl Copy, Alt Link, Shift Move) would require a reviewed Windows backend supporting `DROPEFFECT_MOVE|COPY|LINK`, `CFSTR_PREFERREDDROPEFFECT`, `CFSTR_PERFORMEDDROPEFFECT`, and `CFSTR_LOGICALPERFORMEDDROPEFFECT`, together with reliable validation of move ownership and cancellation. **Not implemented or certified here.** The issue author explicitly permitted continuing if this is not possible. Do not advertise native Move while this backend remains copy-only.

## Validation checklist (not yet witnessed in this environment)

- Build `organizer.exe` and optional `organizer-popup.exe`; run `go test ./...` in the root and helper modules.
- On Windows 11, verify popup position on primary and secondary displays with taskbar on different edges and 100%/150%/200% DPI.
- Drag a large file and Unicode/space-containing filename to Explorer and Desktop, confirm copied bytes, source untouched; press Esc and reject the drag, verify no mutation.
- Choose a shortcut then click each action. Confirm exactly one action applies, verify collision-safe moves, and confirm no other file is affected during simultaneous downloads.
- Remove or rename the helper, verify native fallback; set `provider: native` and `enabled: false`, verify old behavior and suppression respectively.
- Linux X11/Wayland: validate separate desktop sessions using the optional GTK3 helper; native `notify-send` remains the fallback if the helper cannot start. No real-session verification has been performed here.

## Licensing

Original OrganizerV2 sources remain MIT. The optional Wails helper links `eaxum/wails-dragout`, licensed AGPL-3.0. Distributing that helper requires meeting applicable obligations for the combined work, including attribution, preserving exact license text, notices and corresponding source. Switching to the native provider in YAML does **not** remove AGPL from a linked binary. See `LICENSES/AGPL-3.0.txt`, `LICENSES/AGPL-NOTICE.txt` and `THIRD_PARTY_NOTICES.md`. The adapted drag-rs MIT license is also included at `LICENSES/LICENSE.drag-rs`. See the [technical license audit](issue-17-license-audit.md) for outstanding source-offer and dependency-inventory work; obtain formal legal review before distributing combined artifacts.

## Issue #19: popup presentation and fallback

- The Windows helper is linked as a GUI-subsystem PE executable (`-H windowsgui`), so it does not create a console window. Both CI and tag releases verify the PE subsystem in PowerShell.
- The helper creates a hidden WebView, renders all UI controls from Wails bindings, positions it within the primary monitor's **logical WorkArea**, then shows it and sends authenticated READY.
- The launcher only falls back to a native toast for **pre-READY** failures (missing helper, launch error, early exit, timeout, capacity). Once READY has been received, normal window closure or any subsequent helper exit does not generate a duplicate notification.
- Wails bindings and embedded frontend routes are checked in helper tests; generate bindings before `go test ./...` in the helper module.
- The custom popup uses the existing shortcut IDs and native file-action handler, and releases its event when the popup process ends. The watcher remains asynchronous.
- **Manual Windows 11 acceptance is still required** for physical drag-out, DPI/multi-monitor window placement, simultaneous downloads and actual click/move actions. Unit tests and CI cannot certify these desktop interactions.
