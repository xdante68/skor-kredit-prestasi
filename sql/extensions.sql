CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TYPE achievement_status_enum AS ENUM ('draft', 'submitted', 'verified', 'rejected', 'deleted');
-- Index
CREATE INDEX idx_achievements_student_id ON achievement_references(student_id);
CREATE INDEX idx_achievements_status ON achievement_references(status);
CREATE INDEX idx_achievements_created_at ON achievement_references(created_at DESC);
CREATE INDEX idx_achievements_student_status ON achievement_references(student_id, status);
-- Index
CREATE INDEX IF NOT EXISTS idx_students_user_id ON students(user_id);
CREATE INDEX IF NOT EXISTS idx_students_advisor_id ON students(advisor_id);
CREATE INDEX IF NOT EXISTS idx_lecturers_user_id ON lecturers(user_id);