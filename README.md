# ob

This project is not necessarily the best way to synchronize Obsidian vaults with GitHub, but it is the method that works best for me.
While there are many Obsidian plugins designed for this purpose, I wanted a simple command-line tool that I could configure once and then forget about.

I manage multiple vaults, and installing or configuring a synchronization plugin for each of them quickly becomes cumbersome.
The goal of this project is to create a lightweight, reliable CLI tool that can fully replace [Obsidian Sync](https://obsidian.md/sync) for users who prefer or need a free, open-source alternative, particularly students who already rely on GitHub to keep their notes synchronized across devices.

Any contributions are highly welcome! Whether it’s fixing bugs, improving automation, or enhancing usability, every effort helps make this tool more stable and accessible for everyone.

## Getting Started

This project requires **Go version 1.25.4 or higher**. Make sure you have a compatible version installed. If needed, download the latest version from [https://go.dev/dl/](https://go.dev/dl/)

1. **Installation**: Installs ob in the system

    ```bash
    go install github.com/albertoboccolini/ob@latest
    ```

2. **Start ob**: Starts synchronization in the background

    ```bash
    ob start <vault_path>
    ```

## Other useful ob commands

- **Show ob status**: Displays the sync status, whether boot is enabled, and other useful information.

    ```
    ob status
    ```

- **Show the logs**: Displays all logs produced by ob directly in the terminal.

    ```
    ob logs
    ```

- **Enable or disable ob at startup**: Controls whether the system service automatically starts **ob** at boot.

    ```
    ob boot <enable|disable>
    ```

- **Manual sync**: Triggers an immediate synchronization, bypassing the local commit threshold and pushing changes directly to the remote repository.

    ```
    ob sync
    ```

## Vault Requirements

Since `ob` does not yet provide an initialization command to automate repository setup, your vault must meet the following prerequisites before synchronization can begin:

- The vault must already be an initialized Git repository
- A remote must be configured
- Authentication with the remote must already be set up
- The working branch must exist both locally and on the remote

These requirements ensure that `ob` can immediately start synchronizing your vault without additional configuration steps.

## Project hints

- **Uninstall**: Removes ob from your system.

    ```bash
    rm -rf ~/go/bin/ob
    ```

- **Compile** (for development purposes only)

    ```bash
    go build
    ```

- **Lint code** (for development purposes only)

    ```bash
    go fmt ./...
    ```
