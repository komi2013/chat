# Test script to simulate an incoming FCM message via ADB
# This triggers the onMessageReceived in FirebaseMessagingService

$packageName = "com.chat.android"

Write-Host "Simulating Push Notification for $packageName..." -ForegroundColor Cyan

# Use ADB to send a broadcast that simulates an FCM message
# Note: In modern Android, you can't easily trigger the Firebase service directly via broadcast for security.
# This command tests the UI and Notification Manager logic specifically.

# Data-only message: "payload" is the JSON array the server sends via
# FCMManager.SendData; FirebaseMessagingService dispatches on pd[1].
adb shell am broadcast -a com.google.android.c2dm.intent.RECEIVE `
    -n "$packageName/com.chat.android.firebase.FirebaseMessagingService" `
    --es "from" "123456789" `
    --es "pushID" "test-push-1" `
    --es "payload" '["test-push-1","thread","channel-1","tester"]'

Write-Host "If the app is running and permissions are granted, you should see a notification." -ForegroundColor Green
