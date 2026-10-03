-- 000008_mentor_sessions_and_chat.down.sql
-- GitWise Phase 7 Rollback: Drop Mentor Sessions and Chat Messages

DROP TABLE IF EXISTS chat_messages CASCADE;
DROP TABLE IF EXISTS mentor_sessions CASCADE;
