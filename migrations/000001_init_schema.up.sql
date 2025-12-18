-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    google_id VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    credits INTEGER DEFAULT 5 NOT NULL,
    tier VARCHAR(20) DEFAULT 'free' NOT NULL,
    subscription_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);

-- Create templates table
CREATE TABLE templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    category VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    thumbnail_url TEXT NOT NULL,
    preview_video_url TEXT,
    base_prompt TEXT NOT NULL,
    default_params JSONB NOT NULL DEFAULT '{}',
    credit_cost INTEGER DEFAULT 1 NOT NULL,
    estimated_time_seconds INTEGER DEFAULT 60 NOT NULL,
    is_premium BOOLEAN DEFAULT FALSE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    preferred_provider VARCHAR(50),
    tags TEXT[] DEFAULT '{}',
    usage_count BIGINT DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Create video_jobs table
CREATE TABLE video_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES templates(id),
    prompt TEXT NOT NULL,
    params JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'pending' NOT NULL,
    progress INTEGER DEFAULT 0 NOT NULL,
    provider VARCHAR(50),
    provider_job_id VARCHAR(255),
    video_url TEXT,
    thumbnail_url TEXT,
    duration_seconds INTEGER,
    credits_charged INTEGER NOT NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

-- Create indexes for users
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_google_id ON users(google_id) WHERE google_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_users_tier ON users(tier) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at DESC) WHERE deleted_at IS NULL;

-- Create indexes for templates
CREATE INDEX idx_templates_category ON templates(category) WHERE is_active = TRUE;
CREATE INDEX idx_templates_is_premium ON templates(is_premium) WHERE is_active = TRUE;
CREATE INDEX idx_templates_usage_count ON templates(usage_count DESC) WHERE is_active = TRUE;
CREATE INDEX idx_templates_created_at ON templates(created_at DESC);

-- Create indexes for video_jobs
CREATE INDEX idx_video_jobs_user_id ON video_jobs(user_id);
CREATE INDEX idx_video_jobs_user_status ON video_jobs(user_id, status);
CREATE INDEX idx_video_jobs_status ON video_jobs(status);
CREATE INDEX idx_video_jobs_created_at ON video_jobs(created_at DESC);
CREATE INDEX idx_video_jobs_provider ON video_jobs(provider) WHERE provider IS NOT NULL;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_templates_updated_at
    BEFORE UPDATE ON templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

