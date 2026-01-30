# Screenshots for JSMon CLI README

Add screenshots here so the main README displays them. Use the filenames below so links work.

## Suggested screenshots

| Filename | What to capture | Suggested command |
|----------|------------------|--------------------|
| `help.png` | CLI help output (logo + usage) | `jsmon -h` or `jsmon -help` |
| `quickstart.png` | Create workspace + one scan | `jsmon -cw "Test" -key KEY` then `jsmon -u "https://example.com/a.js" -wksp ID` |
| `install.png` | Successful install (terminal) | `go install github.com/jsmonhq/jsmon-cli@latest` and `jsmon -h` |
| `config.png` | Config in use (e.g. listing workspaces) | `jsmon -workspaces -key YOUR_KEY` |
| `scanning.png` | URL or domain scan in progress / result | `jsmon -u "https://example.com/app.js" -wksp ID` |
| `recon.png` | Recon or filter output (JSON) | `jsmon -recon "field=emails page=1" -wksp ID` or `jsmon -filters "urls=github page=1" -wksp ID` |

## Tips

- **Terminal:** Use a readable theme (e.g. dark background, good font size).
- **Crop:** Trim to the relevant part (no need for full desktop).
- **Privacy:** Redact API keys and real workspace IDs if you share the repo publicly.
- **Format:** PNG is fine; keep file size reasonable (e.g. under 500 KB per image).

After adding the files, the main README will show them automatically.
