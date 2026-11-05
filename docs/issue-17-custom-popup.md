# OrganizerV2 issue #17 — production integration

## Current behavior

- `notifications.provider: custom` (default) launches `organizer-popup.exe` next to the headless Windows `organizer.exe`. It uses Wails v3 and `eaxum/wails-dragout` for a **real native copy-only drag**.
- `notifications.provider: native` uses the existing Windows toast unchanged.
- If the helper is missing, cannot start, does not send READY within 10s, or four popups are already active, the organizer logs an error and delivers a native toast. The watcher never waits on the UI thread.
- Existing Open File, Open Location and Confirm actions are routed to the **same** in-memory registry, selector resolver and file-action handler as native toasts, over an ephemeral authenticated 127.0.0.1 endpoint. The UI sees opaque destination IDs, never arbitrary target paths. The registry Claims each event exactly once.
- The popup is frameless, topmost and hidden from the Windows taskbar, positioned within the primary monitor's work area (which excludes the taskbar). Its duration, button visibility and drag availability come from YAML.
- Linux keeps the original `notify-send` notifier, whether `custom` or `native` is selected, until platform-specific native drag-out has been verified.

## Build the optional helper (Windows 11)

```powershell
cd cmd/organizer-popup
# Requires Go 1.27.2+, MinGW GCC, WebView2, and Wails 3 beta.32.
go mod tidy
go run github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.32 generate bindings -d ./assets/bindings
go build -trimpath -tags production -o ../../dist/organizer-popup.exe .
cd ../..
# Keep organizer.exe and organizer-popup.exe side-by-side.
```

The main `go.mod` remains Go 1.25, with `CGO_ENABLED=0`; the popup lives in its own module requiring CGO and Go 1.27.2. Builds that package only the watcher may be distributed without the popup, with native notification fallback.

## Safe drag-out semantics and known limitation

The current upstream `wails-dragout` is **copy-only**. A `dropped` result reports that the drop target accepted a drag, **not** that a file was copied successfully and not that a move occurred. The OrganizerV2 source path is never deleted or updated based on this return value. No Move/Copy/Link modifier claim is made.

The requirement added in issue comment #6646734242 (default Move, Ctrl Copy, Alt Link, Shift Move) would require a reviewed Windows backend supporting `DROPEFFECT_MOVE|COPY|LINK`, `CFSTR_PREFERREDDROPEFFECT`, `CFSTR_PERFORMEDDROPEFFECT`, and `CFSTR_LOGICALPERFORMEDDROPEFFECT`, together with reliable validation of move ownership and cancellation. **Not implemented or certified here.** The issue author explicitly permitted continuing if this is not possible. Do not advertise native Move while this backend remains copy-only.

## Validation checklist (not yet witnessed in this environment)

- Build `organizer.exe` and optional `organizer-popup.exe`; run `go test ./...` in the root and helper modules.
- On Windows 11, verify popup position on primary and secondary displays with taskbar on different edges and 100%/150%/200% DPI.
- Drag a large file and Unicode/space-containing filename to Explorer and Desktop, confirm copied bytes, source untouched; press Esc and reject the drag, verify no mutation.
- Choose a shortcut then click each action. Confirm exactly one action applies, verify collision-safe moves, and confirm no other file is affected during simultaneous downloads.
- Remove or rename the helper, verify native fallback; set `provider: native` and `enabled: false`, verify old behavior and suppression respectively.
- Linux X11/Wayland: validate separate environments; current implementation intentionally uses native notification fallback.

## Licensing

Original OrganizerV2 sources remain MIT. The optional Wails helper links `eaxum/wails-dragout`, licensed AGPL-3.0. Distributing that helper requires meeting applicable obligations for the combined work, including attribution, preserving exact license text, notices and corresponding source. Switching to the native provider in YAML does **not** remove AGPL from a linked binary. See `LICENSES/AGPL-3.0.txt`, `LICENSES/AGPL-NOTICE.txt` and `THIRD_PARTY_NOTICES.md`. Get a legal review before releasing combined artifacts.
