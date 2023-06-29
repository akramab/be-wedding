CREATE TABLE IF NOT EXISTS invitation_sessions(
  id TEXT PRIMARY KEY,
  session_name TEXT NOT NULL,
  schedule TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO invitation_sessions (id, session_name, schedule, created_at)
VALUES
  ('ccd52961-fa4e-43ba-a6df-a4c97849d899', 'SESSION_1', '10.30 - 12.00', '2023-06-29T12:54:18.610Z'),
  ('ccd52961-fa4e-43ba-a6df-a4c97849d898', 'SESSION_2', '12.00 - 13.30', '2023-06-29T12:54:18.610Z');