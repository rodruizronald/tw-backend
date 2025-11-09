// Package main provides a utility to populate the database with technology information.
// It reads technology data and inserts them into the database.
package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/rodruizronald/tw-backend/internal/config"
	"github.com/rodruizronald/tw-backend/internal/database"
	"github.com/rodruizronald/tw-backend/internal/techalias"
	"github.com/rodruizronald/tw-backend/internal/technology"
)

// Technology represents a technology entity as stored in the configuration.
// It contains the basic information needed to create a technology record in the database
type Technology struct {
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Alias    []string `json:"alias"`
	Parent   string   `json:"parent"`
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
	// Configure logger
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	/// Load configuration
	cfg, err := config.Load("../../.env")
	if err != nil {
		log.Errorf("failed to load configuration: %v", err)
		return err
	}

	// Connect to the database using config
	dbpool, err := database.Connect(ctx, &cfg.Database)
	if err != nil {
		log.Errorf("Unable to connect to database: %v", err)
		return err
	}
	defer dbpool.Close()

	// Create repositories
	techRepo := technology.NewRepository(dbpool)
	aliasRepo := techalias.NewRepository(dbpool)

	// Process technologies
	processTechnologies(ctx, log, techRepo, aliasRepo)

	log.Info("Technology import completed")
	return nil
}

// processTechnologies handles the two-pass technology import process
func processTechnologies(ctx context.Context, log *logrus.Logger, techRepo *technology.Repository,
	aliasRepo *techalias.Repository) {
	// Create a map to store all technologies by name for lookup
	techMap := make(map[string]*technology.Technology)

	// Process and insert all technologies
	technologies := readTechnologiesFromJSON()
	log.Infof("Loaded %d technologies from JSON file", len(technologies))

	// First pass: create technologies without parent references
	log.Info("Starting first pass: creating technologies without parent references")
	createTechnologies(ctx, log, techRepo, aliasRepo, technologies, techMap)

	// Second pass: update technologies with parent references
	log.Info("Starting second pass: updating technologies with parent references")
	updateTechnologyParents(ctx, log, techRepo, technologies, techMap)
}

// createTechnologies handles the first pass of creating technologies
func createTechnologies(ctx context.Context, log *logrus.Logger, techRepo *technology.Repository,
	aliasRepo *techalias.Repository, technologies []Technology, techMap map[string]*technology.Technology) {

	for _, tech := range technologies {
		// Create the technology model
		newTech := &technology.Technology{
			Name:     tech.Name,
			Category: tech.Category,
			// Parent ID will be set in the second pass
		}

		var existingTech *technology.Technology

		// Insert into database
		err := techRepo.Create(ctx, newTech)
		if err != nil {
			// Skip if it's a duplicate
			if technology.IsDuplicate(err) {
				log.Infof("Technology already exists: %s", tech.Name)

				// Fetch the existing technology to use for parent mapping
				existingTech, err = techRepo.GetByName(ctx, tech.Name)
				if err != nil {
					log.Warnf("Error fetching existing technology %s: %v", tech.Name, err)
					continue
				}
				techMap[tech.Name] = existingTech

				// Add aliases for existing technology
				addAliases(ctx, log, aliasRepo, existingTech.ID, tech.Alias)
				continue
			}
			log.Warnf("Error creating technology %s: %v", tech.Name, err)
			continue
		}

		log.Infof("Created technology: %s (ID: %d)", tech.Name, newTech.ID)
		techMap[tech.Name] = newTech

		// Add aliases for new technology
		addAliases(ctx, log, aliasRepo, newTech.ID, tech.Alias)
	}
}

// updateTechnologyParents handles the second pass of updating parent references
func updateTechnologyParents(ctx context.Context, log *logrus.Logger, techRepo *technology.Repository,
	technologies []Technology, techMap map[string]*technology.Technology) {

	for _, tech := range technologies {
		if tech.Parent == "" {
			continue // Skip technologies without parents
		}

		// Look up the current technology
		currentTech, exists := techMap[tech.Name]
		if !exists {
			log.Warnf("Cannot find technology: %s", tech.Name)
			continue
		}

		// Look up the parent technology
		parentTech, exists := techMap[tech.Parent]
		if !exists {
			log.Warnf("Cannot find parent technology: %s for %s", tech.Parent, tech.Name)
			continue
		}

		// Update the parent ID
		currentTech.ParentID = &parentTech.ID
		err := techRepo.Update(ctx, currentTech)
		if err != nil {
			log.Warnf("Error updating parent for %s: %v", currentTech.Name, err)
			continue
		}

		log.Infof("Updated technology %s with parent %s (ID: %d)",
			currentTech.Name, parentTech.Name, parentTech.ID)
	}
}

// addAliases adds aliases for a technology
func addAliases(ctx context.Context, log *logrus.Logger, aliasRepo *techalias.Repository,
	techID int, aliases []string) {
	for _, aliasName := range aliases {
		if aliasName == "" {
			continue
		}

		// Create alias model
		newAlias := &techalias.TechnologyAlias{
			TechnologyID: techID,
			Alias:        aliasName,
		}

		// Insert into database
		err := aliasRepo.Create(ctx, newAlias)
		if err != nil {
			// Skip if it's a duplicate
			if techalias.IsDuplicate(err) {
				log.Infof("Alias already exists: %s", aliasName)
				continue
			}
			log.Warnf("Error creating alias %s for technology ID %d: %v", aliasName, techID, err)
			continue
		}

		log.Infof("Created alias: %s (ID: %d) for technology ID %d", aliasName, newAlias.ID, techID)
	}
}

// readTechnologiesFromJSON reads technology data from a JSON file
func readTechnologiesFromJSON() []Technology {
	// Get the directory of the current executable
	execDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logrus.Errorf("Failed to get executable directory: %v", err)
		return []Technology{}
	}

	// Path to the JSON file
	jsonPath := filepath.Join(execDir, "technologies.json")

	// For development, if the file doesn't exist in the executable directory,
	// try looking in the current directory
	if _, err = os.Stat(jsonPath); os.IsNotExist(err) {
		jsonPath = "technologies.json"
	}

	// Read the JSON file
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		logrus.Errorf("Failed to read technologies file: %v", err)
		return []Technology{}
	}

	// Parse the JSON data
	var technologies []Technology
	if err = json.Unmarshal(data, &technologies); err != nil {
		logrus.Errorf("Failed to parse technologies JSON: %v", err)
		return []Technology{}
	}

	return technologies
}
