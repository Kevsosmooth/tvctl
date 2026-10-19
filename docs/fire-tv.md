# Fire TV setup

Fire OS refuses adb connections that start on the TV itself (Amazon change, 2023;
confirmed on Fire OS 7.7.1.7, 2026-10-18). So on a Fire TV the bridge runs on the
customer's phone, which reaches the TV's adb over their Wi-Fi like any outside
device. Shield and Google TV allow on-device adb, so there the TV app alone is enough.

## What the customer does, once

On the Fire TV:

1. If Developer options is missing: Settings > My Fire TV > About, press the
   device name 7 times.
2. Settings > My Fire TV > Developer options > ADB debugging > On.
3. The first time the phone connects, tick "Always allow" and press OK on the
   "Allow USB debugging?" popup.

Step-by-step with screenshots: [setup.md](setup.md#fire-tv).

On the phone: install the app, stay on the same Wi-Fi, share the code.

Nothing is installed on the TV, and "install unknown apps" is not needed: adb
installs skip it. Both settings survive reboots.

## What the helper gets

Full adb from anywhere (installs, updates, settings, keys, typing, screenshots),
but only while the phone app is running on the TV's Wi-Fi. When the phone leaves
or the app stops, the helper is out.

The customer can stop it by closing the phone app, or turn it off for good with
ADB debugging > Off.

## Tested 2026-10-18

Helper on a VPS, bridge on a LAN PC standing in for the phone, Fire TV Cube
(3rd gen, AFTGAZL): linked directly in 11 s with no relay server, screenshot in 0.7 s,
silent reinstall of a 9.8 MB app in 14 s (about 2 MB/s).
