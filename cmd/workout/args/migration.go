package args

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	dbfiles "git.rpjosh.de/RPJosh/workout/db"
	"git.rpjosh.de/RPJosh/workout/internal/api"
	"git.rpjosh.de/RPJosh/workout/internal/dbutils"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"github.com/RPJoshL/go-logger"
)

var (
	lastPartRegex     = regexp.MustCompile(`(\d)*(.*)`)
	sqlDelimiterRegex = regexp.MustCompile(`(?m)^\s*DELIMITER\s+(\S*)$`)
)

type migrationFile struct {
	content   string
	name      string
	version   string
	extension string
}

type Migration struct {
	config  *models.AppConfig
	version string
}

func (m *Migration) SetMigration(cli *Cli) string {
	db := dbutils.New(api.GetDb(&m.config.Db))

	toVersion := padVersion(m.version)

	var currentVersion sql.NullString
	verErr := db.QueryForValue(&currentVersion, "SELECT `release` FROM version ORDER BY update_time, `release` DESC LIMIT 1")

	if isCriticalVersionError(verErr) {
		logger.Fatal("Failed to get current migration version: %s", verErr)
	}

	var err error
	if !currentVersion.Valid {
		err = m.executeMigrations(db, func(paddedVersion string) bool {
			return paddedVersion <= toVersion
		})
	} else {
		fromVersion := padVersion(currentVersion.String)

		err = m.executeMigrations(db, func(paddedVersion string) bool {
			return paddedVersion > fromVersion && paddedVersion <= toVersion
		})
	}

	if err != nil {
		logger.Fatal("Failed to run migrations: %s", err)
	}

	logger.Info("Migrations to version %s completed successfully", m.version)
	os.Exit(0)

	return ""
}

// isCriticalVersionError checks if the error is critical for
// running the migration if the current version could not be
// determined from the db.
//
// We don't want to execute an incorrect migration when the
// db is gone away but still need to accept errors in the following cases:
//   - No rows in version table
//   - Missing version table
func isCriticalVersionError(err database.Error) bool {
	if err == nil {
		return false
	}

	if err.Type() == database.NoRows {
		return false
	}

	msg := err.Error()

	// Table does not exist
	if strings.Contains(msg, "Error 1146 (42S02)") {
		return false
	}

	return true
}

// padVersion pads a version string to ensure proper lexicographical comparison
func padVersion(version string) (rtc string) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		logger.Fatal("Invalid version format: %q", version)
	}

	for _, part := range parts {
		if rtc != "" {
			rtc += "."
		}

		for len(part) < 4 {
			part = "0" + part
		}

		if len(part) > 4 {
			logger.Fatal("Invalid version part length in version: %s", version)
		}

		rtc += part
	}

	return rtc
}

func (m *Migration) executeMigrations(db *dbutils.Db, shouldApply func(paddedVersion string) bool) error {
	files, err := dbfiles.Migrations.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("reading migration files: %w", err)
	}

	// Get all executions to run
	migrationFiles := []*migrationFile{}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		parts := strings.Split(file.Name(), ".")
		if len(parts) != 4 {
			continue
		}

		// The last part may contain an extension (when multiple migration files are available)
		matches := lastPartRegex.FindStringSubmatch(parts[2])
		if len(matches) != 3 {
			logger.Debug("Invalid version part with possible suffix: %s", parts[2])
			continue
		}

		version := padVersion(strings.Join([]string{parts[0], parts[1], matches[1]}, "."))
		if !shouldApply(version) {
			continue
		}

		content, err := dbfiles.Migrations.ReadFile("migrations/" + file.Name())
		if err != nil {
			return fmt.Errorf("reading embedded file: %w", err)
		}

		migrationFiles = append(migrationFiles, &migrationFile{
			name:      file.Name(),
			version:   version,
			extension: matches[2],
			content:   string(content),
		})
	}

	// Sort all files to run migrations in correct order
	slices.SortStableFunc(migrationFiles, func(a, b *migrationFile) int {
		return strings.Compare(a.version+a.extension, b.version+b.extension)
	})

	for idx, mig := range migrationFiles {
		if err := runMigration(mig, db); err != nil {
			return fmt.Errorf("executing migration %s: %w", mig.name, err)
		}

		// Insert version to mark it as successfully migrated
		if idx+1 >= len(migrationFiles) || migrationFiles[idx+1].version != mig.version {
			if err := insertVersion(mig.version, db); err != nil {
				return fmt.Errorf("inserting version (%s) after successful migration: %w", mig.version, err)
			}
		}
	}

	return nil
}

func runMigration(mig *migrationFile, db *dbutils.Db) error {
	logger.Info("Running migration for version %s%s", mig.version, mig.extension)

	stmts := getStatements(mig.content)
	for idx, content := range stmts {
		if len(stmts) > 1 {
			logger.Debug("Executing statement %d for migration %s%s", idx+1, mig.version, mig.extension)
		}

		if _, err := db.Db.Exec(content); err != nil {
			return fmt.Errorf("snippet with index %d: %w", idx, err)
		}
	}
	return nil
}

func getStatements(content string) (rtc []string) {
	if !strings.HasPrefix(strings.TrimSpace(content), "DELIMITER") {
		return []string{content}
	}

	sqlDelimiterMatches := sqlDelimiterRegex.FindStringSubmatch(content)
	if len(sqlDelimiterMatches) != 2 {
		logger.Fatal("Invalid regex result. Got %d results", len(sqlDelimiterMatches))
	}

	delimiter := sqlDelimiterMatches[1]

	// Remove all delimiter statements. They are not valid SQL
	content = sqlDelimiterRegex.ReplaceAllString(content, "")

	for mig := range strings.SplitSeq(content, delimiter) {
		migTrimmed := strings.TrimSpace(mig)
		if migTrimmed != "" {
			rtc = append(rtc, migTrimmed)
		}
	}

	return rtc
}

func insertVersion(version string, db *dbutils.Db) error {
	_, err := db.Db.Exec("INSERT INTO version(`release`) VALUES (?)", version)
	return err
}
