CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TYPE achievement_status_enum AS ENUM ('draft', 'submitted', 'verified', 'rejected', 'deleted');