ALTER TABLE users ADD COLUMN IF
    NOT EXISTS is_vip BOOLEAN NOT NULL DEFAULT FALSE;