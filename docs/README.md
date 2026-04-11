# Chat Application Documentation

This directory contains comprehensive documentation for the chat application.

## Available Documentation

### [Limits and Restrictions](./limits-and-restrictions.md)
Complete overview of all system limits including:
- Session limits (5 browser/device sessions)
- Channel limits (3 channels can be created)
- User profile limits (3 nicknames)
- File upload limits
- Posting restrictions
- Data retention policies

## Code Structure

### Backend (Go)
- `controller/` - HTTP request handlers and business logic
- `common/` - Shared utilities and helper functions
- `collection/` - Data structures and models
- `console/` - Background tasks and cleanup jobs
- `infrastructure/` - Database and external service configurations

### Frontend (Vue.js)
- `vue/src/components/` - Reusable Vue components
- `vue/src/views/` - Page-level Vue components
- `vue/src/pushReceive/` - Push notification handling
- `vue/src/store/` - State management

### Database
- MongoDB for primary data storage
- Collections: session, user, channel, alias, etc.

## Key Features

- Multi-channel chat system
- Real-time push notifications
- File sharing capabilities
- User management with nicknames
- Mobile-responsive design
- Google authentication integration

## Development

The application uses:
- **Backend**: Go with MongoDB driver
- **Frontend**: Vue.js 3 with modern UI components
- **Authentication**: Google OAuth 2.0
- **Real-time**: Web push notifications
- **File Storage**: Local filesystem with cleanup

For detailed technical specifications, see the individual documentation files.
