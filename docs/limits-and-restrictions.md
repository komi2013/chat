# Limits and Restrictions Documentation

This document outlines all the system limits and restrictions for the chat application.

## Session Limits

### Browser/Device Sessions
- **Maximum**: 5 concurrent sessions per user across all browsers/devices
- **Location**: `controller/SignInGoogle.go` lines 203-210
- **Behavior**:
  - Same browser: Old session is replaced when signing in again
  - Different browsers: Each creates a new session up to the 5-session limit
  - Sessions older than 1 month are automatically deleted

### Session Duration
- **Cookie Max Age**: 30 days (2592000 seconds)
- **Session Expiration**: 20 days of inactivity
- **Session Regeneration**: After 10 days of inactivity

## Channel Limits

### Channel Creation
- **Maximum**: 3 channels per user that can be created as owner/admin
- **Location**: `controller/ChannelAdd.go` lines 48-51
- **Note**: This limit only applies to **creating** new channels, not joining existing ones

### Channel Membership
- **Maximum**: Unlimited - users can join as many existing channels as they want
- **Implementation**: Each channel membership creates a `ChannelAlias` entry

## User Profile Limits

### Nicknames
- **Maximum**: 3 nicknames per user
- **Location**: `controller/UserEdit.go` lines 103-105
- **Frontend**: Vue component also enforces `MAX_NICKNAMES = 3`
- **Purpose**: Profile nicknames with custom images/emojis

### ChannelAliases
- **Maximum**: No explicit limit
- **Purpose**: Records which channels a user belongs to and their alias name in each channel
- **Behavior**: Created automatically when joining channels

## Content Limits

### File Uploads
- **Maximum file size**: 100MB per file
- **Maximum total size**: 500MB per upload
- **Maximum file count**: 10 files per upload
- **Location**: `common/file.go` lines 105-108

### Push Contents
- **Maximum**: 100 items in session push queue
- **Location**: `common/session.go` lines 107-113

### Posting Limits
- **Thread limit**: 3 threads within 20 hours
- **Post limit**: 3 posts per thread within 20 hours
- **Location**: `controller/TweetPost.go` lines 60-76

## Data Retention

### File Cleanup
- **ContentsPush files**: 5 days (120 hours)
- **Tweet files**: 1 hour
- **Default files**: 24 hours
- **Location**: `console/FileClean.go` lines 41-50

### Message Limits
- **Tweet pagination**: 10 messages per page
- **Latest tweets**: 100 messages maximum
- **Location**: `controller/TweetGet.go` and `controller/TweetGetLatest.go`

## Security Limits

### Push Notification Chunks
- **Maximum chunk size**: 2000 bytes
- **Location**: `common/push.go` lines 73-80

### String Lengths
- **Display names**: 30 characters maximum (with truncation)
- **Location**: `common/string.go` lines 118-121

## Configuration

All these limits are hardcoded in the application source code. To modify any limit:

1. Locate the relevant file mentioned above
2. Update the numeric value
3. Rebuild and redeploy the application

## Related Files

- `controller/SignInGoogle.go` - Session management
- `controller/ChannelAdd.go` - Channel creation limits
- `controller/UserEdit.go` - Nickname limits
- `controller/TweetPost.go` - Posting limits
- `common/session.go` - Session handling
- `common/file.go` - File upload limits
- `console/FileClean.go` - Data retention policies
