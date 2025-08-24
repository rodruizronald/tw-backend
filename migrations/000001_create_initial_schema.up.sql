-- Companies Table
CREATE TABLE companies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Jobs Table (Modified: search_vector is now a regular column)
CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    company_id INT REFERENCES companies(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    responsibilities TEXT[],
    skill_must_have TEXT[],
    skill_nice_have TEXT[],
    main_technologies TEXT[],
    benefits TEXT[],
    experience_level VARCHAR(50) NOT NULL, 
    employment_type VARCHAR(50) NOT NULL,
    location VARCHAR(50) NOT NULL,
    work_mode VARCHAR(20) NOT NULL,
    application_url VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    signature VARCHAR(64) UNIQUE,
    search_vector tsvector, -- Regular column now, not generated
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Function to update search_vector
CREATE OR REPLACE FUNCTION update_jobs_search_vector()
RETURNS trigger AS $$
BEGIN
    NEW.search_vector := 
        setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(array_to_string(NEW.skill_must_have, ' '), '')), 'B') ||
        setweight(to_tsvector('english', coalesce(array_to_string(NEW.skill_nice_have, ' '), '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to maintain search_vector on INSERT and UPDATE
CREATE TRIGGER update_jobs_search_vector_trigger
BEFORE INSERT OR UPDATE OF title, description, skill_must_have, skill_nice_have
ON jobs
FOR EACH ROW
EXECUTE FUNCTION update_jobs_search_vector();

-- Technologies Table (canonical names)
CREATE TABLE technologies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(50) NOT NULL,
    parent_id INT REFERENCES technologies(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Technology Aliases Table
CREATE TABLE technology_aliases (
    id SERIAL PRIMARY KEY,
    technology_id INT NOT NULL REFERENCES technologies(id) ON DELETE CASCADE,
    alias VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Job Technologies Junction Table
CREATE TABLE job_technologies (
    id SERIAL PRIMARY KEY,
    job_id INT REFERENCES jobs(id) ON DELETE CASCADE,
    technology_id INT REFERENCES technologies(id) ON DELETE CASCADE,
    is_required BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(job_id, technology_id)
);

-- Create Indexes
-- Companies Indexes
CREATE UNIQUE INDEX idx_companies_name ON companies(name);
CREATE INDEX idx_companies_active ON companies(id) WHERE is_active = TRUE;

-- Jobs Indexes
-- Create a GIN index on the search_vector column to speed up full-text search queries
-- GIN (Generalized Inverted Index) is optimized for full-text search
CREATE INDEX idx_jobs_search_vector ON jobs USING GIN (search_vector);
CREATE INDEX idx_jobs_location ON jobs(location);
CREATE INDEX idx_jobs_active ON jobs(id) WHERE is_active = TRUE;
CREATE INDEX idx_jobs_work_mode ON jobs(work_mode);
CREATE INDEX idx_jobs_created_at ON jobs(created_at);
CREATE INDEX idx_jobs_company_id ON jobs(company_id);
CREATE UNIQUE INDEX idx_jobs_signature ON jobs(signature);
CREATE INDEX idx_jobs_employment_type ON jobs(employment_type);
CREATE INDEX idx_jobs_experience_level ON jobs(experience_level);

-- Technologies Indexes
CREATE UNIQUE INDEX idx_technologies_name_lower ON technologies(LOWER(name));
CREATE INDEX idx_technologies_category ON technologies(category);
CREATE INDEX idx_technologies_parent_id ON technologies(parent_id);

-- Technology Aliases Indexes
CREATE INDEX idx_technology_aliases_technology_id ON technology_aliases(technology_id);
CREATE INDEX idx_technology_aliases_alias_lower ON technology_aliases(LOWER(alias));

-- Job Technologies Indexes
CREATE INDEX idx_job_technologies_job_id ON job_technologies(job_id);
CREATE INDEX idx_job_technologies_technology_id ON job_technologies(technology_id);