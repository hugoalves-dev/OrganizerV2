# Issue #17: acceptance matrix and evidence

> **Status:** source-level and CI checks are automated; physical drag/drop is **not** certified by CI. Never mark Windows 11, X11, or native Wayland drag-out as passed without a recorded desktop run.

## Prerequisites

- Build both binaries **for the same platform and architecture** and put the watcher and helper side-by-side.
- Windows: WebView2, `organizer.exe` and `organizer-popup.exe`. Linux amd64: GTK3/WebKit2GTK 4.1, `organizer-linux-amd64` and executable `organizer-popup`.
- Linux should be tested twice: `echo "$XDG_SESSION_TYPE"` must report `x11` and `wayland` in their separate sessions. An XWayland application on a Wayland desktop does **not** prove native Wayland DnD.
- Set `notifications.provider: custom`, `enabled: true`, `custom.drag_and_drop.enabled: true`. Use test-only watched/destination folders, never personal documents.
- Keep `provider: native` as fallback. On Linux without an installed/working helper, `notify-send` is required to see the fallback.
- Record binary SHA-256, distro/Windows build, compositor, desktop/file manager, monitor/DPI, GTK/Wayland backend (`GDK_BACKEND` if set), Wails version, actual drop result and screenshots/logs.

## Automated coverage (GitHub CI)

| Area | Code/evidence |
|---|---|
| Watcher: Windows/Linux unit tests, race detector and vet | `go test -race ./...` / `go vet ./...` |
| Optional helper Windows build + tests | `go test ./...`, `go build -tags production` |
| Optional Linux amd64 GTK3 helper build + tests | `go test -tags gtk3 ./...`, `go build -tags 'gtk3 production'` |
| Missing helper / capacity / closed launcher | `popup_lifecycle_test.go` |
| Linux action routing, Unicode and collision-safe move/copy | `file_action_service_linux_test.go` |
| No silent source deletion on drag cancel | `PopupService.StartDrag` only passes file path to copy-only upstream; neither action callback nor drag result deletes it |
| Required MIT/AGPL notices tracked | `license_notice_test.go`; both build bundles copy `LICENSES/*` |

Automated tests cannot synthesize a trusted file-manager drag source, Escape release, window focus, Wayland input serials, monitor scaling or actual byte-for-byte cross-volume drag result.

## Manual desktop acceptance (record PASS/FAIL per environment)

Run every applicable check in **Windows 11 Explorer/Desktop**, **Linux X11 file manager**, and **native Linux Wayland file manager**. Keep a saved pre-test checksum for the source.

| ID | Action | Expected |
|---|---|---|
| A01 | Organize one file | Frameless popup at bottom right, above usable taskbar/panel, file name/category shown |
| A02 | Put mouse over popup for longer than configured timeout; leave | Stays visible while hovered, then dismisses; watcher remains responsive |
| A03 | Drag to desktop and file-manager folder | Real OS file copy, correct bytes, original preserved; `dropped` is not a checksum guarantee |
| A04 | Drag and press Escape/reject target | No source mutation, no unwanted Open/Confirm action |
| A05 | Drag filename with accents, Japanese characters and spaces | Correct bytes/name, no broken URI encoding |
| A06 | Drag a large file, compare SHA-256 and size at destination | Identical destination content; source preserved |
| A07 | Drop on existing filename | No unreported overwrite; record file-manager collision behavior separately from OrganizerV2's shortcut moves |
| A08 | Select a configured Move file to shortcut, click Open Location | File moves once using collision-safe naming; correct destination opened |
| A09 | Repeat shortcut with Open File | File moves once and newly moved file opens |
| A10 | Select shortcut and Confirm | File moves once, no extra Open action |
| A11 | Disable each action and `move_to` in YAML | Disabled buttons/shortcuts are not available |
| A12 | Generate 5 quick downloads | No swapped IDs, source paths or shortcuts; beyond 4 concurrent popups uses native fallback |
| A13 | Rename/delete file before dragging | Graceful error, no accidental delete/move |
| A14 | Remove helper, use `provider: native`, and set `enabled: false` | Native fallback; native-only; no notification, respectively |
| A15 | DPI 100%, 150%, 200%, second monitor, taskbar/panel other edges | Popup placement/visibility recorded; unsupported setups documented |
| A16 | Ctrl, Alt, Shift during drag | Current upstream is **copy-only**, NOT Move/Link negotiation; document actual file-manager behavior without claiming full modifier support |
| A17 | Denied/unsupported drop target | Error/cancelled without deleting source or consuming an action |
| A18 | Compare Linux X11 and native Wayland | Record whether pointer/serial requirements allow drag; do not infer from build success |

### Manual pass criteria

- Windows: A01–A14 and A17 should pass before changing the PR from draft. A15–A16 must have results/limitations documented.
- Linux X11 and Wayland: A01–A14 and A17–A18 must be tested **independently** before saying native Linux drag-out is supported. Until confirmed, treat the helper as experimental and retain `notify-send`.
- For A03/A06, compare actual file checksums; a native `dropped` return only means the target accepted the drag.
- **No automatic deletion of source after drag-out**. A future negotiated Move backend needs its own dedicated acceptance criteria.
- A reviewer must confirm `LICENSES/AGPL-3.0.txt`, `LICENSES/AGPL-NOTICE.txt`, `LICENSES/LICENSE.drag-rs`, `LICENSES/MIT.txt`, `THIRD_PARTY_NOTICES.md`, and corresponding source availability before AGPL helper releases.

## Manual report template

```text
Date and tester:
Environment: Windows 11 / Linux X11 / Linux native Wayland
OS build / desktop / file manager / GTK backend:
Wails helper SHA256:
Checks: A01..A18 = PASS | FAIL | NOT APPLICABLE
Problem + reproducible steps + log:
Source SHA256 before and after; destination SHA256:
Evidence links:
```

Do not treat a CI badge, source review, or an XWayland test as proof that physical Wayland drag-and-drop passed.
