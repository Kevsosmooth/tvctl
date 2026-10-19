# Setup scripts

`tvctl run <script.yaml> -g <group>` plays the same steps on every chosen TV at once
(`--parallel`, default 8). Each TV stops at its first failed step and reports
`N/M steps`. Other TVs keep going.

## Steps

Each step has exactly one action.

| Step | Example | What it does |
|---|---|---|
| `key` | `key: [home, down, ok]` | Press remote keys in order (`tvctl keys` lists names) |
| `text` | `text: "my password"` | Type into the focused field |
| `click` | `click: "Install"` | Find on-screen text and press it |
| `wait_for` | `wait_for: "Sign in"` | Wait until the text shows (default 20s, change with `timeout: 40s`) |
| `wait` | `wait: 2s` | Pause |
| `launch` | `launch: com.stremio.one` | Open an app (TV launcher entry first) |
| `stop` | `stop: com.stremio.one` | Force-stop an app |
| `install` | `install: apks/stremio.apk` | Install or update an APK |
| `uninstall` | `uninstall: com.example.bloat` | Remove an app |
| `shell` | `shell: settings put global stay_on_while_plugged_in 0` | Run an adb shell command |
| `screenshot` | `screenshot: shots/{device}-done.png` | Save a 1280px-wide PNG |

Paths are relative to the script's folder. `{device}` becomes the TV's name.

## Example: install Stremio and open it

```yaml
name: stremio
steps:
  - key: [wake, home]
  - install: apks/stremio.apk
  - launch: com.stremio.one
  - wait_for: "Log in"
    timeout: 40s
  - screenshot: shots/{device}-stremio.png
  - key: [home]
```

```
tvctl run stremio.yaml -g customer-smith
tvctl run stremio.yaml --all
```

## Tips for agents

- Run `tvctl ui -t <tv>` to see what is on screen before writing `click` and
  `wait_for` steps. Labels are matched case-insensitively. An exact match beats a
  partial one.
- `tvctl screenshot -t <tv>` shows the screen when `ui` comes back empty. Fire TV
  sometimes refuses UI dumps while an Amazon overlay is up.
- A TV screen that is asleep has no UI. Start scripts with `key: [wake]`.
- Logins that need a code from a phone (TV sign-in pages) cannot be scripted. Stop
  the script there, screenshot the code, and finish by hand.
