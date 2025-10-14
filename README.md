# ob

This project is not necessarily the best way to synchronize Obsidian vaults with GitHub, but it is the method that works best for me.
While there are many Obsidian plugins designed for this purpose, I wanted a simple command-line tool that I could configure once and then forget about.

I manage multiple vaults, and installing or configuring a synchronization plugin for each of them quickly becomes cumbersome.
The goal of this project is to create a lightweight, reliable CLI tool that can fully replace [Obsidian Sync](https://obsidian.md/sync) for users who prefer or need a free, open-source alternative, particularly students who already rely on GitHub to keep their notes synchronized across devices.

Any contributions are highly welcome! Whether it’s fixing bugs, improving automation, or enhancing usability, every effort helps make this tool more stable and accessible for everyone.

## Getting Started

This project requires **Go version 1.25.2 or higher**. Make sure you have a compatible version installed. If needed, download the latest version from [https://go.dev/dl/](https://go.dev/dl/)

1. **Installation**: Installs ob in the system

    ```bash
    go install
    ```

2. **Start ob**: Starts synchronization in the background

    ```bash
    ob start <vault_path>
    ```

## Additional

- **Stop ob**: Stops the synchronization process.

    ```
    ob stop
    ```

- **Uninstall**: Removes ob from your system.

    ```bash
    rm -rf ~/go/bin/ob
    ```

- **Compile** (for development purpouses only)

    ```bash
    go build
    ```

- **Lint code** (for development purpouses only)

    ```bash
    go fmt ./...
    ```
