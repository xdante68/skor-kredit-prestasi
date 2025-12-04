CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TYPE achievement_status_enum AS ENUM ('draft', 'submitted', 'verified', 'rejected', 'deleted');
-- Index
CREATE INDEX idx_achievements_student_id ON achievement_references(student_id);
CREATE INDEX idx_achievements_status ON achievement_references(status);
CREATE INDEX idx_achievements_created_at ON achievement_references(created_at DESC);
CREATE INDEX idx_achievements_student_status ON achievement_references(student_id, status);