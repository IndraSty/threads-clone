-- migrations/001_create_users_table.sql
-- Migration to create users table with OAuth support

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create Users table with OAuth support
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255), -- Can be NULL for OAuth-only users
    bio TEXT,
    profile_image_url VARCHAR(500),
    oauth_providers JSONB, -- Store OAuth provider data
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_oauth_google ON users USING GIN ((oauth_providers->'google'));
CREATE INDEX IF NOT EXISTS idx_users_oauth_facebook ON users USING GIN ((oauth_providers->'facebook'));
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Insert some sample data for testing (with hashed passwords)
-- Password for all test users: "password123"
INSERT INTO users (username, display_name, email, password_hash, bio) VALUES
('john_doe', 'John Doe', 'john@example.com', '$2a$10$K7L/fnoeE/5jHN5.JfzrVeq1JZ3aULvf.7BpfBNYGK7LFvD5QGT7m', 'Software developer and tech enthusiast'),
('jane_smith', 'Jane Smith', 'jane@example.com', '$2a$10$K7L/fnoeE/5jHN5.JfzrVeq1JZ3aULvf.7BpfBNYGK7LFvD5QGT7m', 'Designer and creative thinker'),
('bob_wilson', 'Bob Wilson', 'bob@example.com', '$2a$10$K7L/fnoeE/5jHN5.JfzrVeq1JZ3aULvf.7BpfBNYGK7LFvD5QGT7m', 'Product manager at a tech startup')
ON CONFLICT (email) DO NOTHING;

-- Example OAuth user (Google login)
INSERT INTO users (
    username, 
    display_name, 
    email, 
    oauth_providers, 
    profile_image_url
) VALUES (
    'alice_google',
    'Alice Johnson', 
    'alice.google@example.com',
    '{"google": {"id": "123456789", "email": "alice.google@example.com"}}'::jsonb,
    'https://lh3.googleusercontent.com/a/default-user'
) ON CONFLICT (email) DO NOTHING;

-- Example OAuth user (Facebook login)
INSERT INTO users (
    username, 
    display_name, 
    email, 
    oauth_providers,
    profile_image_url
) VALUES (
    'charlie_facebook',
    'Charlie Brown', 
    'charlie.facebook@example.com',
    '{"facebook": {"id": "987654321", "email": "charlie.facebook@example.com"}}'::jsonb,
    'https://graph.facebook.com/987654321/picture'
) ON CONFLICT (email) DO NOTHING;

-- Example user with both OAuth providers linked
INSERT INTO users (
    username, 
    display_name, 
    email, 
    password_hash,
    oauth_providers,
    bio
) VALUES (
    'david_multi',
    'David Multi', 
    'david.multi@example.com',
    '$2a$10$K7L/fnoeE/5jHN5.JfzrVeq1JZ3aULvf.7BpfBNYGK7LFvD5QGT7m',
    '{"google": {"id": "111111111", "email": "david.multi@example.com"}, "facebook": {"id": "222222222", "email": "david.multi@example.com"}}'::jsonb,
    'Full-stack developer who loves both Google and Facebook'
) ON CONFLICT (email) DO NOTHING;