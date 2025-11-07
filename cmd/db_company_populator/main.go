// Package main provides a utility to populate the database with company information.
// It reads from a JSON file of companies and inserts them into the database.
package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/rodruizronald/ticos-in-tech/internal/company"
	"github.com/rodruizronald/ticos-in-tech/internal/config"
	"github.com/rodruizronald/ticos-in-tech/internal/database"
)

// Company represents a company entity as stored in the JSON configuration file.
// It contains the basic information needed to create a company record in the database.
type Company struct {
	Name string `json:"name"`
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

	// Read companies from JSON file
	companies, err := readCompaniesFromJSON()
	if err != nil {
		log.Errorf("Failed to read companies from JSON: %v", err)
		return err
	}
	log.Infof("Loaded %d companies from JSON file", len(companies))

	// Load configuration
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

	// Create a company repository
	repo := company.NewRepository(dbpool)

	// Store each company in the database
	for _, c := range companies {
		cm := &company.Company{
			Name:     c.Name,
			IsActive: true,
		}

		err = repo.Create(ctx, cm)
		if err != nil {
			if company.IsDuplicate(err) {
				log.Infof("Company already exists: %s", cm.Name)
				continue
			}
			log.Warnf("Error creating company %s: %v", c.Name, err)
			continue
		}

		log.Infof("Successfully added company: %s (ID: %d)", cm.Name, cm.ID)
	}

	log.Info("Company population completed")
	return nil
}

// readCompaniesFromJSON reads the companies data from a JSON file
func readCompaniesFromJSON() ([]Company, error) {
	// Get the directory of the current executable
	execDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}

	// Path to the JSON file
	jsonPath := filepath.Join(execDir, "companies.json")

	// For development, if the file doesn't exist in the executable directory,
	// try looking in the current directory
	if _, err = os.Stat(jsonPath); os.IsNotExist(err) {
		jsonPath = "companies.json"
	}

	// Read the JSON file
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	// Parse the JSON data
	var companies []Company
	if err := json.Unmarshal(data, &companies); err != nil {
		return nil, err
	}

	return companies, nil
}
