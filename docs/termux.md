# Running hacelo on the phone with Termux (experimental)

The server runs on the same Android phone that receives the reminders, so no
hosting is needed. The app is opened at `http://localhost:8080`, which Chrome
treats as a secure context (service worker and push work without HTTPS).

## 1. Build the binary (on your computer)

Requires Go 1.22+ and Node 20+.

```bash
./scripts/build-android.sh
```

This produces `dist/hacelo-arm64`. Generate the VAPID keys once with
`cd backend && go run ./cmd/genvapid` and keep them.

## 2. Set up Termux (on the phone)

1. Install Termux and Termux:Boot from F-Droid (not from Google Play).
2. In Termux:

```bash
pkg update && pkg install proot
termux-setup-storage
cp ~/storage/downloads/hacelo-arm64 ~/hacelo && chmod +x ~/hacelo
```

Copy `dist/hacelo-arm64` to the phone's Downloads folder first (USB cable,
Drive, etc.).

`proot` is needed because a static Go binary on Android cannot resolve DNS by
itself (no `/etc/resolv.conf`). `termux-chroot` provides it.

## 3. Run it

```bash
termux-wake-lock
export PORT=8080
export DB_PATH=$HOME/hacelo.db
export API_TOKEN=choose-a-long-random-string
export VAPID_PUBLIC_KEY=...
export VAPID_PRIVATE_KEY=...
export VAPID_SUBSCRIBER=mailto:you@example.com
export SSL_CERT_FILE=$PREFIX/etc/tls/cert.pem
termux-chroot ~/hacelo
```

Then open `http://localhost:8080` in Chrome, use **Add to Home screen**, open
**Ajustes**, paste the `API_TOKEN` and tap **Activar notificaciones**.

## 4. Keep it alive

Android aggressively kills background apps, which is the main risk of this
setup:

- Settings -> Apps -> Termux -> Battery -> **Unrestricted** (same for Chrome).
- Keep `termux-wake-lock` active (the notification stays in the shade).
- Optional: put the commands above in `~/.termux/boot/hacelo.sh` so the server
  starts after a reboot (requires Termux:Boot to be opened once).

## Known caveats

- Data lives only on the phone (`~/hacelo.db`). Back it up if it matters.
- If Android kills Termux, reminders stop silently until it is restarted.
- This setup is untested; if it proves unreliable, deploy the same Docker image
  to a free VM or Fly.io instead.
