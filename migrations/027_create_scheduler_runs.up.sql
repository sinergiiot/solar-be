-- Migration 027: Create scheduler_runs table for visibility
CREATE TABLE IF NOT EXISTS scheduler_runs (
    id SERIAL PRIMARY KEY,
    job_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL, -- 'success', 'failed'
    duration_ms BIGINT NOT NULL,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    finished_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_scheduler_runs_job_name ON scheduler_runs(job_name);
CREATE INDEX IF NOT EXISTS idx_scheduler_runs_finished_at ON scheduler_runs(finished_at);
