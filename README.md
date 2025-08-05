# Notification2Alert

**Notification2Alert** is a minimalist cross-platform utility that monitors a specified folder and instantly displays a system notification for every new file with a special xattr attribute.

## 🛠 Features

- Watches a directory and all subfolders for new files
- Checks xattr attribute: `type=notification`
- Reads additional attributes: `from`, `summary`
- Displays a native system notification:
    - Title: `from`
    - Message: `summary`
    - Clicking the notification opens the related file (supported on macOS and most desktop Linux)
- Cross-platform: macOS, Linux, Windows
    - macOS: custom icon support (with proper bundle or sender spoofing)
    - macOS: reliable notification click-to-open via terminal-notifier or gosx-notifier
- Sends a test notification on launch for verification

## 🔧 Installation

1. Build the binary:
    ```bash
    go build
    ```

2. Create a `config.json` file next to the binary:
    ```json
    {
      "folder": "/path/to/watched/folder"
    }
    ```

## 🚀 How It Works

1. Loads config and watches the given folder (and subfolders).
2. For every new file with `type=notification`:
    - Reads the `from` and `summary` attributes
    - Shows a system notification with those fields
    - Clicking the notification opens the file (when supported)
3. Works on all major desktop OS (using beeep, gosx-notifier, and/or terminal-notifier)
4. Displays a test notification on startup

## ⚠️ Notes

- On macOS, for full icon/click support, install [terminal-notifier](https://github.com/julienXX/terminal-notifier):
    ```bash
    brew install terminal-notifier
    ```
- Some notification features (icons, click handlers) depend on OS and configuration.
- To spoof sender and change icon on macOS, you may set the `Sender` field to a system app's bundle ID (e.g., Safari).

## License

MIT
