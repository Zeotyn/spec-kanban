## MODIFIED Requirements

### Requirement: Plugin builds on install
The plugin SHALL provide its `kanban` binary during `herdr plugin install`. When the installed commit is a release whose binary for the machine's OS and architecture can be downloaded and verified against the release checksums, the install SHALL use that binary and SHALL NOT need a Go toolchain. In every other case the plugin SHALL build the binary from source, and the installation SHALL fail with herdr's build error when the binary cannot be built. For a linked development checkout, the repository SHALL document a single command that builds the binary at the path the manifest uses.

#### Scenario: Install of a release without Go
- **WHEN** a user runs `herdr plugin install zeotyn/openspec-kanban --ref v<version>` for a published release on a supported OS and architecture, on a machine without a Go toolchain
- **THEN** the install succeeds with the downloaded, checksum-verified binary and `kanban --version` prints `<version>`

#### Scenario: Install from GitHub with Go available
- **WHEN** a user runs `herdr plugin install zeotyn/openspec-kanban` on a machine with a Go toolchain
- **THEN** the install succeeds and the plugin's actions can be invoked without further steps

#### Scenario: Install without Go
- **WHEN** a user installs a commit that has no matching release binary (or the download or verification fails) on a machine without a Go toolchain
- **THEN** the install aborts, the plugin is not registered, and herdr shows the failed build command

#### Scenario: Linked checkout
- **WHEN** a developer links the checkout with `herdr plugin link` and runs the documented build command
- **THEN** invoking the open action launches the freshly built binary
