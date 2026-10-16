# Android App Links & Deferred Invitations Setup

This document outlines the steps to configure and verify App Links and Install Referrer handling for the Chat application.

## 1. App Links Verification (assetlinks.json)

To enable automatic verification for `https://chat.quigen.info`, you must host a Digital Asset Links file on your web server.

**Location:** `https://chat.quigen.info/.well-known/assetlinks.json`

**Content:**
```json
[
  {
    "relation": ["delegate_permission/common.handle_all_urls"],
    "target": {
      "namespace": "android_app",
      "package_name": "com.chat.android",
      "sha256_cert_fingerprints": [
        "5C:86:50:79:4D:BB:1C:53:3F:1A:EE:C2:A5:4B:6B:DB:4C:61:92:E6:F0:D6:93:D7:E8:9A:EF:D8:4C:30:CD:C4"
      ]
    }
  }
]
```

### How to get your SHA256 fingerprint:
- **Debug:** Run `./gradlew signingReport` and look for the SHA256 under `debug` variant.
- **Release:** If using Play App Signing, get the fingerprint from the Google Play Console under **Setup > App Integrity**.

## 2. Supported URL Formats

The app handles the following URL patterns:
- `https://chat.quigen.info/invitation/{M}/?code={invitationCode}`
- `https://chat.quigen.info/profile/{M}/?code={invitationCode}`
- `chat://profile/{M}?code={invitationCode}`

## 3. Deferred Invitations (Install Referrer)

When a user clicks an invitation link but doesn't have the app installed, they are directed to the Play Store. After installation, the app retrieves the referrer string to complete the invitation.

**Expected Referrer Format:** `m={M}&code={invitationCode}`

**Test with ADB:**
```bash
adb shell am broadcast -a com.android.vending.INSTALL_REFERRER -n com.chat.android/com.chat.android.core.receiver.InstallReferrerReceiver --es "referrer" "m=shop001&code=abc123"
```
*(Note: Modern versions of Android/Play Store handle this internally; for testing, the `InstallReferrerClient` in `MainActivity` is the primary entry point.)*

## 4. Verification

Use the App Links Assistant in Android Studio (**Tools > App Links Assistant**) to:
1. Verify Intent Filters.
2. Verify the `assetlinks.json` file on the server.
3. Test the URL mapping.
