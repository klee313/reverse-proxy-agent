# rpa-android

Minimal Android UI shell for rpa client features.

## What is included
- Status screen (start/stop, status summary)
- Logs screen (searchable recent logs)
- Config editor (YAML text editor)
- Metrics screen (key/value list)
- Doctor screen (diagnostic checklist)
- Foreground service stub (status bar notification)
- SSH public key view (copy/share)

## Run (Android Studio)
1. Open `apps/rpa-android` in Android Studio.
2. Sync Gradle and run `app` on a device/emulator.

## Notes
- This is UI-only scaffolding. Service/SSH integration will be wired next.
- Gradle wrapper is included (`./gradlew`).
