-- Drop indexes first
DROP INDEX IF EXISTS idx_job_technologies_technology_id;
DROP INDEX IF EXISTS idx_job_technologies_job_id;

DROP INDEX IF EXISTS idx_technology_aliases_alias_lower;
DROP INDEX IF EXISTS idx_technology_aliases_technology_id;

DROP INDEX IF EXISTS idx_technologies_parent_id;
DROP INDEX IF EXISTS idx_technologies_category;
DROP INDEX IF EXISTS idx_technologies_name_lower;

DROP INDEX IF EXISTS idx_jobs_search_vector;
DROP INDEX IF EXISTS idx_jobs_experience_level;
DROP INDEX IF EXISTS idx_jobs_employment_type;
DROP INDEX IF EXISTS idx_jobs_signature;
DROP INDEX IF EXISTS idx_jobs_company_id;
DROP INDEX IF EXISTS idx_jobs_created_at;
DROP INDEX IF EXISTS idx_jobs_work_mode;
DROP INDEX IF EXISTS idx_jobs_active;
DROP INDEX IF EXISTS idx_jobs_location;

DROP INDEX IF EXISTS idx_companies_active;
DROP INDEX IF EXISTS idx_companies_name;

-- Drop trigger first (before dropping the function it uses)
DROP TRIGGER IF EXISTS update_jobs_search_vector_trigger ON jobs;

-- Drop function
DROP FUNCTION IF EXISTS update_jobs_search_vector();

-- Drop tables in reverse order of creation (to handle dependencies)
DROP TABLE IF EXISTS job_technologies;
DROP TABLE IF EXISTS technology_aliases;
DROP TABLE IF EXISTS technologies;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS companies;
