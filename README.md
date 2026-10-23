# Notification2Alert

**Notification2Alert** is a minimalist cross-platform utility that monitors a specified folder and instantly displays a system notification for every new file with a special xattr attribute.

## 🛠 Features

- Watches a directory and all subfolders for new files
- Checks xattr attribute: `type=notification`
- Reads additional attributes: `from`, `summary`
- Resolves the sender's email against optional `.vcf` contact cards, displaying the
  contact or company name (and the contact photo on macOS, when present)
- Displays a native system notification:
    - Title: contact/company name from VCF, or `from` when no card matches
    - Message: `summary`
    - Clicking the notification opens the related file (supported on macOS and most desktop Linux)
- Cross-platform: macOS, Linux, Windows
    - macOS: custom icon support (with proper bundle or sender spoofing)
    - macOS: reliable notification click-to-open via terminal-notifier or gosx-notifier

## 🔧 Installation

1. Build the binary:
    ```bash
    go build
    ```

   Keep `AppIcon.icns` next to the built binary. On macOS it is used as the
   notification application icon.

2. Create a `config.json` file next to the binary:
    ```json
    {
      "folder": "/path/to/watched/folder",
      "contacts": "/path/to/vcf/folder"
    }
    ```
    `contacts` is optional: a folder with `.vcf` cards. If the `from` email belongs to a card,
    the notification title shows the contact/company name (email goes to the subtitle)
    and the card photo is attached as the notification image.

## 🚀 How It Works

1. Loads config and watches the given folder (and subfolders).
2. For every new file with `type=notification`:
    - Reads the `from` and `summary` attributes
    - Looks up `from` in the optional VCF contacts directory
    - Shows a system notification using the matched contact/company name, or the sender email when no match is found
    - Clicking the notification opens the file (when supported)
3. Works on all major desktop OS (using beeep, gosx-notifier, and/or terminal-notifier)

## ⚠️ Notes

- On macOS, for full icon/click support, install [terminal-notifier](https://github.com/julienXX/terminal-notifier):
    ```bash
    brew install terminal-notifier
    ```
- Some notification features (icons, click handlers) depend on OS and configuration.
- To spoof sender and change icon on macOS, you may set the `Sender` field to a system app's bundle ID (e.g., Safari).

## License

MIT
