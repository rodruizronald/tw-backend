// Package main provides a utility to populate the database with job information.
// It reads job data from JSON files and inserts them into the database along with
// their associated technologies.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/rodruizronald/tw-backend/internal/company"
	"github.com/rodruizronald/tw-backend/internal/config"
	"github.com/rodruizronald/tw-backend/internal/database"
	"github.com/rodruizronald/tw-backend/internal/jobs"
	"github.com/rodruizronald/tw-backend/internal/jobtech"
	"github.com/rodruizronald/tw-backend/internal/techalias"
	"github.com/rodruizronald/tw-backend/internal/technology"
)

// Job define a type to represent a single job
type jobData struct {
	Company          string   `json:"company"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Responsibilities []string `json:"responsibilities"`
	Requirements     struct {
		MustHave   []string `json:"must_have"`
		NiceToHave []string `json:"nice_to_have"`
	} `json:"requirements"`
	MainTechnologies []string `json:"main_technologies"`
	Benefits         []string `json:"benefits"`
	ApplicationURL   string   `json:"application_url"`
	Location         string   `json:"location"`
	WorkMode         string   `json:"work_mode"`
	ExperienceLevel  string   `json:"experience_level"`
	EmploymentType   string   `json:"employment_type"`
	Technologies     []struct {
		Name     string `json:"name"`
		Category string `json:"category"`
		Required bool   `json:"required"`
	} `json:"technologies"`
	Signature string `json:"signature"`
}

// Update the jobs struct to use the Job type
type internalJobs struct {
	Jobs []jobData `json:"jobs"`
}

// RemovedSignatures represents the structure of the removed_signatures.json file
type RemovedSignatures struct {
	RemovedSignatures []string  `json:"removed_signatures"`
	Count             int       `json:"count"`
	Timestamp         time.Time `json:"timestamp"`
	Reason            string    `json:"reason"`
}

func main() {
	var err error
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		stop()
		if err != nil {
			os.Exit(1)
		}
	}()
	err = run(ctx)
}

func run(ctx context.Context) error {
	// Initialize logger
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Setup database and repositories
	dbpool, repos, err := setupDatabase(ctx, log)
	if err != nil {
		return err
	}
	defer dbpool.Close()

	// Get file paths
	today := time.Now().Format("20060102")
	inputDir := filepath.Join("data", today)
	inputFile := filepath.Join(inputDir, "jobs.json")
	missingTechFile := filepath.Join(inputDir, "missing_technologies.json")
	removedSignaturesFile := filepath.Join(inputDir, "removed_signatures.json")

	// Process removed signatures first
	if processErr := processRemovedSignatures(ctx, removedSignaturesFile, repos, log); processErr != nil {
		return processErr
	}

	// Read and parse job data
	jobData, err := readJobData(inputFile, log)
	if err != nil {
		return err
	}

	// Process jobs and collect missing technologies
	missingTechnologies, err := processJobs(ctx, jobData, repos, log)
	if err != nil {
		return err
	}

	// Write missing technologies to file if any
	if err := writeMissingTechnologies(missingTechnologies, missingTechFile, log); err != nil {
		return err
	}

	log.Info("Job population completed")
	return nil
}

// setupDatabase initializes the database connection and repositories
func setupDatabase(ctx context.Context, log *logrus.Logger) (*pgxpool.Pool, *repositories, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Errorf("failed to load configuration: %v", err)
		return nil, nil, err
	}

	// Connect to the database using config
	dbpool, err := database.Connect(ctx, &cfg.Database)
	if err != nil {
		log.Errorf("Unable to connect to database: %v", err)
		return nil, nil, err
	}

	// Create repositories
	repos := &repositories{
		job:     jobs.NewRepository(dbpool),
		company: company.NewRepository(dbpool),
		jobtech: jobtech.NewRepository(dbpool),
		tech:    technology.NewRepository(dbpool),
		alias:   techalias.NewRepository(dbpool),
	}

	return dbpool, repos, nil
}

// repositories holds all the database repositories needed
type repositories struct {
	job     *jobs.Repository
	company *company.Repository
	jobtech *jobtech.Repository
	tech    *technology.Repository
	alias   *techalias.Repository
}

// readJobData reads and parses the job data from the input file
func readJobData(inputFile string, log *logrus.Logger) (*internalJobs, error) {
	log.Infof("Reading job data from %s", inputFile)

	// Read job data from file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		log.Errorf("Failed to read job data file: %v", err)
		return nil, err
	}

	// Parse job data
	var jobData internalJobs
	if err := json.Unmarshal(data, &jobData); err != nil {
		log.Errorf("Failed to parse job data: %v", err)
		return nil, err
	}

	log.Infof("Found %d jobs to process", len(jobData.Jobs))
	return &jobData, nil
}

// processJobs processes each job and returns a map of missing technologies
func processJobs(ctx context.Context, jobData *internalJobs, repos *repositories,
	log *logrus.Logger) (map[string][]string, error) {
	// Create a map to track missing technologies
	missingTechnologies := make(map[string][]string) // company -> list of missing tech names

	// Process each job
	for i := range jobData.Jobs {
		j := &jobData.Jobs[i] // Use a pointer to the job instead of copying it

		// Process job and its technologies
		jobMissingTechs, err := processJob(ctx, j, repos, log)
		if err != nil {
			// Log error but continue with next job
			log.Warnf("Error processing job %s: %v", j.Title, err)
			continue
		}

		// Add any missing technologies to the map
		if len(jobMissingTechs) > 0 {
			missingTechnologies[j.Company] = append(missingTechnologies[j.Company], jobMissingTechs...)
		}
	}

	return missingTechnologies, nil
}

// Update the processJob function signature
func processJob(ctx context.Context, j *jobData, repos *repositories, log *logrus.Logger) ([]string, error) {
	// Find company by name
	jobCompany, err := repos.company.GetByName(ctx, j.Company)
	if err != nil {
		log.Warnf("Error finding company %s: %v", j.Company, err)
		return nil, err
	}

	companyID := jobCompany.ID

	// Create job model
	jobModel := &jobs.Job{
		CompanyID:        companyID,
		Title:            j.Title,
		Description:      j.Description,
		Responsibilities: j.Responsibilities,
		SkillMustHave:    j.Requirements.MustHave,
		SkillNiceHave:    j.Requirements.NiceToHave,
		MainTechnologies: j.MainTechnologies,
		Benefits:         j.Benefits,
		ExperienceLevel:  j.ExperienceLevel,
		EmploymentType:   j.EmploymentType,
		Location:         j.Location,
		WorkMode:         j.WorkMode,
		ApplicationURL:   j.ApplicationURL,
		IsActive:         true,
		Signature:        j.Signature,
	}
	log.Infof("Processing job: %s at %s", jobModel.Title, j.Company)

	// Insert or retrieve job
	if err := createOrRetrieveJob(ctx, jobModel, j, repos.job, log); err != nil {
		return nil, err
	}

	log.Infof("Successfully added job: %s at %s (ID: %d)",
		jobModel.Title, j.Company, jobModel.ID)

	// Process technologies for this job
	return processTechnologies(ctx, j, jobModel, repos, log)
}

// createOrRetrieveJob creates a new job or retrieves an existing one
func createOrRetrieveJob(ctx context.Context, jobModel *jobs.Job, j *jobData, jobRepo *jobs.Repository,
	log *logrus.Logger) error {
	err := jobRepo.Create(ctx, jobModel)
	if err != nil {
		if jobs.IsDuplicate(err) {
			log.Infof("Job already exists: %s at %s", j.Title, j.Company)

			// Get the existing job by signature to retrieve its ID
			existingJob, findErr := jobRepo.GetBySignature(ctx, j.Signature)
			if findErr != nil {
				log.Warnf("Failed to retrieve existing job %s: %v", j.Title, findErr)
				return findErr
			}

			// Use the existing job's ID for technology associations
			jobModel.ID = existingJob.ID
			log.Infof("Using existing job ID: %d", jobModel.ID)
			return nil
		}
		log.Warnf("Failed to insert job %s: %v", j.Title, err)
		return err
	}
	return nil
}

// processTechnologies processes all technologies for a job
func processTechnologies(ctx context.Context, j *jobData, jobModel *jobs.Job, repos *repositories,
	log *logrus.Logger) ([]string, error) {
	var missingTechs []string

	for _, tech := range j.Technologies {
		// Find technology by name or alias
		techModel, err := findTechnology(ctx, tech.Name, repos, log)
		if err != nil {
			missingTechs = append(missingTechs, tech.Name)
			continue
		}

		// Create job technology association
		if err := createJobTechnology(ctx, jobModel.ID, techModel.ID,
			tech.Required, tech.Name, repos.jobtech, log); err != nil {
			continue
		}
	}

	return missingTechs, nil
}

// findTechnology tries to find a technology by name or alias
func findTechnology(ctx context.Context, techName string, repos *repositories,
	log *logrus.Logger) (*technology.Technology, error) {
	// Find technology by name
	techModel, err := repos.tech.GetByName(ctx, techName)
	if err == nil {
		return techModel, nil
	}

	// If not found by exact name, try to find by alias
	alias, aliasErr := repos.alias.GetByAlias(ctx, techName)
	if aliasErr != nil {
		log.Warnf("Technology not found by name or alias: %s: %v", techName, err)
		return nil, aliasErr
	}

	// Get the technology using the alias's technology ID
	techModel, err = repos.tech.GetByID(ctx, alias.TechnologyID)
	if err != nil {
		log.Warnf("Error finding technology by alias ID %d: %v", alias.TechnologyID, err)
		return nil, err
	}

	log.Infof("Found technology %s via alias %s", techModel.Name, techName)
	return techModel, nil
}

// createJobTechnology creates a job-technology association
func createJobTechnology(ctx context.Context, jobID, techID int, isRequired bool, techName string,
	jobtechRepo *jobtech.Repository, log *logrus.Logger) error {
	jobTechModel := &jobtech.JobTechnology{
		JobID:        jobID,
		TechnologyID: techID,
		IsRequired:   isRequired,
	}

	// Insert job technology into database
	err := jobtechRepo.Create(ctx, jobTechModel)
	if err != nil {
		if jobtech.IsDuplicate(err) {
			log.Debugf("Job technology association already exists: %s for job ID %d", techName, jobID)
			return nil
		}
		log.Warnf("Failed to insert job technology %s: %v", techName, err)
		return err
	}

	log.Infof("Added technology %s to job ID %d", techName, jobID)
	return nil
}

// writeMissingTechnologies writes missing technologies to a file
func writeMissingTechnologies(missingTechnologies map[string][]string,
	missingTechFile string, log *logrus.Logger) error {
	if len(missingTechnologies) == 0 {
		return nil
	}

	missingTechData, err := json.MarshalIndent(missingTechnologies, "", "  ")
	if err != nil {
		log.Errorf("Failed to marshal missing technologies: %v", err)
		return err
	}

	err = os.WriteFile(missingTechFile, missingTechData, 0o644)
	if err != nil {
		log.Errorf("Failed to write missing technologies file: %v", err)
		return err
	}

	log.Infof("Missing technologies saved to %s", missingTechFile)
	return nil
}

// processRemovedSignatures reads the removed signatures file and deactivates corresponding jobs
func processRemovedSignatures(
	ctx context.Context,
	removedSignaturesFile string,
	repos *repositories,
	log *logrus.Logger,
) error {
	// Check if the file exists
	if _, err := os.Stat(removedSignaturesFile); os.IsNotExist(err) {
		log.Infof("No removed signatures file found at %s, skipping deactivation", removedSignaturesFile)
		return nil
	}

	log.Infof("Processing removed signatures from %s", removedSignaturesFile)

	// Read the removed signatures file
	data, err := os.ReadFile(removedSignaturesFile)
	if err != nil {
		log.Errorf("Failed to read removed signatures file: %v", err)
		return err
	}

	// Parse the JSON data
	var removedSigs RemovedSignatures
	if err := json.Unmarshal(data, &removedSigs); err != nil {
		log.Errorf("Failed to parse removed signatures JSON: %v", err)
		return err
	}

	log.Infof("Found %d signatures to deactivate", len(removedSigs.RemovedSignatures))

	// Process each signature
	deactivatedCount := 0
	for _, signature := range removedSigs.RemovedSignatures {
		if err := deactivateJobBySignature(ctx, signature, repos.job, log); err != nil {
			log.Warnf("Failed to deactivate job with signature %s: %v", signature, err)
			continue
		}
		deactivatedCount++
	}

	log.Infof("Successfully deactivated %d jobs out of %d signatures", deactivatedCount, len(removedSigs.RemovedSignatures))
	return nil
}

// deactivateJobBySignature finds a job by signature and marks it as inactive
func deactivateJobBySignature(ctx context.Context, signature string, jobRepo *jobs.Repository, log *logrus.Logger) error {
	// Find the job by signature
	job, err := jobRepo.GetBySignature(ctx, signature)
	if err != nil {
		if jobs.IsNotFound(err) {
			log.Debugf("Job with signature %s not found, may have been already removed", signature)
			return nil // Not an error if job doesn't exist
		}
		return fmt.Errorf("failed to find job with signature %s: %w", signature, err)
	}

	// Check if job is already inactive
	if !job.IsActive {
		log.Debugf("Job %s (ID: %d) is already inactive", job.Title, job.ID)
		return nil
	}

	// Mark job as inactive
	job.IsActive = false

	// Update the job in the database
	if err := jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job %s (ID: %d) to inactive: %w", job.Title, job.ID, err)
	}

	log.Infof("Successfully deactivated job: %s (ID: %d, Signature: %s)", job.Title, job.ID, signature)
	return nil
}
