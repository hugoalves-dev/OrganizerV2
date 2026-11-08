# Issue #17: license compliance audit (technical, not legal advice)

## Source inventory reviewed

- OrganizerV2 original: MIT, `LICENSE` and `LICENSES/MIT.txt`.
- `github.com/eaxum/wails-dragout` uses the PR #2 source commit `02b1a12bf3b7` from `github.com/hugoalves-dev/wails-dragout` (pseudo-version `v0.0.0-20261108214754-02b1a12bf3b7`): AGPL-3.0, upstream `LICENSE` and `NOTICE`. `LICENSES/AGPL-3.0.txt` and `LICENSES/AGPL-NOTICE.txt` preserve these texts.
- Adapted `drag-rs` contribution: MIT, Copyright (c) 2023 - Present CrabNebula Ltd. The exact upstream `LICENSE.drag-rs` is copied into `LICENSES/LICENSE.drag-rs` and `THIRD_PARTY_NOTICES.md`.
- Tabler Icons brand asset attribution: MIT, identified in `THIRD_PARTY_NOTICES.md`.
- Other transitive dependencies (Wails, GTK/WebKit, Go packages): **not a completed license-by-license legal audit**. Recheck the dependency manifest and runtime libraries for each release.

## Distribution policy

1. `organizer` remains a separate headless Go binary (no Wails/AGPL drag-out linked to its executable); the helper is optional and built from a separate Go module with CGO.
2. Bundles and releases containing `organizer-popup` include the AGPL/drag-rs/MIT notices, complete corresponding source for the helper and any applicable build/install scripts, plus an appropriate source availability offer. A public repository alone is not proof that every corresponding-source obligation has been met.
3. Document the exact helper commit, dependency versions and build instructions used for release. Ensure recipients can obtain the corresponding source for the **distributed binary version**.
4. An MIT-only distribution excludes the linked AGPL helper and AGPL-derived code; `provider: native` in YAML by itself does **not** exclude it.
5. Keep LGPL/other dependency obligations, dynamic linkage requirements and library notices under review. Running the helper as a separate process does not automatically settle the legal status of the combined distribution.
6. Do not rewrite or merge the original license texts. An actual redistribution/legal release decision requires qualified legal review.

## Release gate

- [x] Preserve project MIT text
- [x] Preserve upstream drag-out AGPL and NOTICE
- [x] Preserve adapted drag-rs original MIT license
- [x] Copy license directory into Windows and Linux helper bundles
- [ ] Verify the **entire** final dependency/license manifest and dynamic libraries
- [ ] Supply and test version-matched Corresponding Source distribution procedure
- [ ] Obtain legal approval for the specific distribution format before shipping AGPL helper

Dependency source used for the Windows alpha: https://github.com/hugoalves-dev/wails-dragout/tree/02b1a12bf3b716e14abbda66ef8ba87793ca9f12 (PR #2, issue #1 remains open for Linux follow-up).

License terms: https://www.gnu.org/licenses/agpl-3.0.html
