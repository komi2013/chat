-- 1. Channel Table
CREATE TABLE channel (
    channel_id VARCHAR(255) PRIMARY KEY,
    channel_name VARCHAR(255) NOT NULL,
    channel_description TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255),
    myname VARCHAR(255) NOT NULL,
    invitation_code VARCHAR(255),
    invitation_guest_code VARCHAR(255),
    invited_at TIMESTAMPTZ
);

-- 2. Alias Table (1-to-Many with Channel)
CREATE TABLE alias (
    alias_id VARCHAR(255) PRIMARY KEY,
    channel_id VARCHAR(255) NOT NULL REFERENCES channel(channel_id) ON DELETE CASCADE,
    alias_name VARCHAR(255),
    alias_img TEXT,
    user_id VARCHAR(255),
    alias_bio TEXT,
    access_right VARCHAR(100)
);

-- 3. Group Table (1-to-Many with Channel)
-- Note: "group" is double-quoted because GROUP is a SQL reserved keyword
CREATE TABLE "group" (
    group_id VARCHAR(255) PRIMARY KEY,
    channel_id VARCHAR(255) NOT NULL REFERENCES channel(channel_id) ON DELETE CASCADE,
    group_name VARCHAR(255),
    group_img TEXT,
    group_bio TEXT,
    alias_names TEXT[] -- Native PostgreSQL array for string slices
);

-- 4. Invitation Table (Optional standalone table linked to Channel)
CREATE TABLE invitation (
    code VARCHAR(255) PRIMARY KEY,
    channel_id VARCHAR(255) NOT NULL REFERENCES channel(channel_id) ON DELETE CASCADE,
    guest_code VARCHAR(255),
    invited_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for Foreign Keys to optimize joins
CREATE INDEX idx_alias_channel_id ON alias(channel_id);
CREATE INDEX idx_group_channel_id ON "group"(channel_id);
CREATE INDEX idx_invitation_channel_id ON invitation(channel_id);


-- File Table
CREATE TABLE file (
    file_path TEXT PRIMARY KEY,
    channel_id VARCHAR(255) REFERENCES channel(channel_id) ON DELETE SET NULL,
    uploaded_by VARCHAR(255),
    public_path TEXT,
    file_size DOUBLE PRECISION,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    available_by TEXT[],
    -- file_type SMALLINT, -- Commented out in Go struct
    usage_type SMALLINT NOT NULL DEFAULT 0
);

-- Index for Foreign Key lookup
CREATE INDEX idx_file_channel_id ON file(channel_id);

-- Column Comments
COMMENT ON COLUMN file.file_path IS 'Primary key representing the storage file path';
COMMENT ON COLUMN file.channel_id IS 'Foreign key referencing channel(channel_id)';
COMMENT ON COLUMN file.uploaded_by IS 'User ID or identifier of the uploader';
COMMENT ON COLUMN file.public_path IS 'Public URL or external access path';
COMMENT ON COLUMN file.file_size IS 'Size of the file in bytes';
COMMENT ON COLUMN file.available_by IS 'List of user IDs or roles granted access';
COMMENT ON COLUMN file.usage_type IS 'Usage purpose: 0=contentsPush, 1=aliasImg, 2=groupImg, 3=nickImg, 4=tweet content, 5=advertisement, 6=reception menu';

-- If file_type is uncommented in the future:
-- COMMENT ON COLUMN file.file_type IS 'File category: 0=other, 1=img, 2=file, 3=icon';

-- 1. User Table
-- Note: "user" is double-quoted because USER is a PostgreSQL reserved keyword
CREATE TABLE "user" (
    user_id VARCHAR(255) PRIMARY KEY,
    google_jwt_sub VARCHAR(255),
    mail VARCHAR(255),
    telephone VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    signed_at TIMESTAMPTZ,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    wallet_address VARCHAR(255)
);

-- Column Comments for User
COMMENT ON TABLE "user" IS 'Stores core user profile and authentication metadata';
COMMENT ON COLUMN "user".user_id IS 'Primary key identifier for the user';
COMMENT ON COLUMN "user".google_jwt_sub IS 'Google OAuth Subject ID (sub claim)';
COMMENT ON COLUMN "user".mail IS 'User email address';
COMMENT ON COLUMN "user".telephone IS 'User telephone number';
COMMENT ON COLUMN "user".signed_at IS 'Timestamp when user last signed terms/login';
COMMENT ON COLUMN "user".latitude IS 'Geographic latitude (e.g., 35.73)';
COMMENT ON COLUMN "user".longitude IS 'Geographic longitude (e.g., 139.53)';
COMMENT ON COLUMN "user".wallet_address IS 'Cryptocurrency/Web3 wallet address';


-- 2. User Channel Alias Table (1-to-Many for User's ChannelAliases)
CREATE TABLE user_channel_alias (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(user_id) ON DELETE CASCADE,
    channel_id VARCHAR(255) NOT NULL REFERENCES channel(channel_id) ON DELETE CASCADE,
    alias VARCHAR(255),
    guest_flag BOOLEAN DEFAULT FALSE
);

COMMENT ON TABLE user_channel_alias IS 'Maps channels and channel aliases associated with a specific user';
COMMENT ON COLUMN user_channel_alias.guest_flag IS 'Indicates whether the user is participating as a guest in this channel';


-- 3. Session Table
CREATE TABLE session (
    session_id VARCHAR(255) PRIMARY KEY,
    device_id VARCHAR(255),
    user_id VARCHAR(255) REFERENCES "user"(user_id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    push_token TEXT,
    device_type SMALLINT DEFAULT 0,
    csrf VARCHAR(255),
    push_contents TEXT[],
    mail VARCHAR(255),
    telephone VARCHAR(50),
    nickname VARCHAR(255),
    nick_img TEXT
);

-- Column Comments for Session
COMMENT ON TABLE session IS 'Active user session state and push notification configuration';
COMMENT ON COLUMN session.device_id IS 'Client-generated UUIDv4 representing the app installation/device. Used for replacing sessions on login and 1-notification-per-device deduplication.';
COMMENT ON COLUMN session.user_id IS 'Foreign key referencing user(user_id)';
COMMENT ON COLUMN session.push_token IS 'Single registered push destination: VAPID JSON (1), FCM token (2), or APNs token (3)';
COMMENT ON COLUMN session.device_type IS 'Session type marker: 0=none, 1=VAPID JSON, 2=FCM token, 3=APNs reserved. Values >= 2 indicate native logins.';
COMMENT ON COLUMN session.csrf IS 'CSRF security token associated with this session';
COMMENT ON COLUMN session.push_contents IS 'Array of pending push notification payloads or IDs';


-- 4. Session Channel Alias Table (1-to-Many for Session's ChannelAliases)
CREATE TABLE session_channel_alias (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL REFERENCES session(session_id) ON DELETE CASCADE,
    channel_id VARCHAR(255) NOT NULL REFERENCES channel(channel_id) ON DELETE CASCADE,
    alias VARCHAR(255),
    guest_flag BOOLEAN DEFAULT FALSE
);

COMMENT ON TABLE session_channel_alias IS 'Active channel aliases attached to a specific session';


-- 5. Session Tweet Post Table (1-to-Many for Session's TweetPosts)
CREATE TABLE session_tweet_post (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL REFERENCES session(session_id) ON DELETE CASCADE,
    parent_id VARCHAR(255),
    post_count INT DEFAULT 0,
    post_admin_flag BOOLEAN DEFAULT FALSE,
    posted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE session_tweet_post IS 'Tracks tweet posting history and counts per session';
COMMENT ON COLUMN session_tweet_post.parent_id IS 'Parent post ID if this is a reply or sub-post';
COMMENT ON COLUMN session_tweet_post.post_admin_flag IS 'Indicates whether the post was created with admin privileges';


-- Indexes for FK performance optimization
CREATE INDEX idx_user_channel_alias_user_id ON user_channel_alias(user_id);
CREATE INDEX idx_session_user_id ON session(user_id);
CREATE INDEX idx_session_channel_alias_session_id ON session_channel_alias(session_id);
CREATE INDEX idx_session_tweet_post_session_id ON session_tweet_post(session_id);

