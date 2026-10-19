# Customer setup

What the TV owner does once, before a helper can connect. It takes about a
minute, survives reboots, and can be undone at any time by turning ADB off.

| Device | Unlock Developer options | Switch to turn on | App goes on |
|---|---|---|---|
| NVIDIA Shield | Device Preferences > About > **Build**, 7 presses | **USB debugging** and **Network debugging** | The TV |
| Google TV / Android TV | System > About > **Android TV OS build**, 7 presses | **USB debugging** | The TV |
| Fire TV (Stick, Cube) | My Fire TV > About > **device name**, 7 presses | **ADB debugging** | A phone on the same Wi-Fi ([why](fire-tv.md)) |

Then, on every device, the first time a helper connects: tick **Always allow
from this computer** and press **OK** (**Allow** on the Shield).

Two different switches, often confused:

- **ADB (Developer options)** is what lets the helper control the TV. Required.
- **Unknown sources / Install unknown apps** only lets an app such as Downloader
  install APKs. It is needed only if the app is sideloaded instead of coming
  from an app store. Once ADB is on, the helper's installs do not need it.

## Fire TV

1. Settings (gear) > **My Fire TV** > **About**.

   ![My Fire TV menu](screenshots/fire-tv/1-my-fire-tv.png)

2. Press the first item, the device name (**Fire TV Stick**, **Fire TV Cube**,
   ...), 7 times, until "You are now a developer" appears.

   ![About menu, device name first](screenshots/fire-tv/2-about.png)

3. Go back to **My Fire TV** > **Developer options** > **ADB debugging** >
   **ON**. (Install unknown apps, below it, is not needed.)

   ![Developer options with ADB debugging on](screenshots/fire-tv/3-developer-options.png)

4. On the phone: install the app, join the same Wi-Fi as the TV, share the code.
5. When the helper first connects, this appears on the TV. Tick **Always allow
   from this computer**, then press **OK**.

   ![Allow USB debugging popup](screenshots/fire-tv/4-allow-debugging.png)

Screenshots from a Fire TV Cube (3rd gen), Fire OS 7.7.1.7. Sticks use the
same menus.

## NVIDIA Shield

1. Settings > **Device Preferences** > **About**, scroll down and press
   **Build** 7 times.

   ![About, scrolled down to Build](screenshots/shield/1-about-build.png)

2. Settings > **Device Preferences** > **Developer options**, scroll down to
   **Debugging** and turn on both **USB debugging** and **Network debugging**.
   Network debugging lets the helper in; USB debugging is what makes the
   approval popup appear for a new helper.

   ![Developer options, Debugging section](screenshots/shield/2-developer-options.png)

3. Install the app on the Shield and share the code.
4. When the helper first connects, tick **Always allow from this computer** and
   press **Allow**.

   ![Allow Network debugging popup](screenshots/shield/3-allow-debugging.png)

Optional, same screen: **Disable adb authorization timeout** on. Otherwise
Android forgets an approved helper who has not connected for 7 days, and the
customer has to press Allow again.

Sideloading the app instead of installing it from Google Play: Settings >
**Device Preferences** > **Security & restrictions** > **Unknown sources**, allow
the downloader app.

![Shield Security & restrictions with Unknown sources](screenshots/shield/unknown-sources.png)

Screenshots from a Shield TV (2019), Android 11, SW 9.1.0. With USB debugging
off, a new helper is refused without any popup (tested 2026-10-19).

## Google TV and other Android TV

1. Settings > **System** > **About**, press **Android TV OS build** 7 times.
2. Settings > **System** > **Developer options** > **USB debugging** > on.
   Turn on **Disable adb authorization timeout** too if it is there (Android 11
   and later forget a helper after 7 days without it).
3. Install the app on the TV and share the code.
4. Tick **Always allow from this computer** and press **OK** when it appears.

Sideloading: Settings > **Apps** > **Security & restrictions** > **Unknown
sources**, allow the downloader app.

Not yet tested on a Google TV device; menu names can differ slightly by brand.

## Turning it off

- Close or stop the app: the helper is disconnected.
- Turn ADB / Network / USB debugging off: no helper can connect until it is
  turned back on.
