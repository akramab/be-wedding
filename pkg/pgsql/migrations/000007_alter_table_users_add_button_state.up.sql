ALTER TABLE users ADD COLUMN IF
    NOT EXISTS is_date_reminder_sent BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF
    NOT EXISTS is_video_reminder_sent BOOLEAN NOT NULL DEFAULT FALSE;