package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"

	"git.rpjosh.de/RPJosh/go-ddl-parser"
	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const City1000 = "./dependencies/cities1000.txt"

func main() {
	defer logger.CloseFile()

	// Get the generic configuration of the app
	conf := models.GetAppConfig()
	api := api.Api{Config: conf}
	db := database.NewDatabaseUtils(api.GetDb())

	// Data to insert
	data := []models.Geonames{}

	// Open file and process every line
	file, err := os.OpenFile(City1000, os.O_RDONLY, os.ModeAppend)
	if err != nil {
		logger.Fatal("Failed to open %q. Did you run 'make install-dependencies'?", City1000)
	}
	defer file.Close()

	// Truncate table
	if _, err := db.Db.GetDb().Exec("TRUNCATE TABLE geonames"); err != nil {
		logger.Fatal("Failed to truncate table data for 'geonames'")
	}

	// Read line with bufio
	r := bufio.NewScanner(file)
	i := 0
	var wg sync.WaitGroup
	for r.Scan() {
		i++
		// Seperated by tab
		vals := strings.Split(r.Text(), "\t")

		d := models.Geonames{
			Geonameid:  toInt(vals[0]),
			Name:       vals[1],
			Location:   ddl.Location{Longitude: toFloat(vals[5]), Latitude: toFloat(vals[4])},
			Country:    vals[8],
			Population: toInt(vals[14]),
		}

		// Alternative names
		if vals[3] != "" {
			d.Alternatenames = database.NewNullString(vals[3])
		}

		data = append(data, d)

		// Insert 10.000 Rows at once. This should have the best performance
		if i%10000 == 0 {
			// Pass by copy
			go func(dd []models.Geonames, index int) {
				defer wg.Done()

				p := message.NewPrinter(language.English)
				if _, err := db.Struct.InsertSlice(&dd).Run(); err != nil {
					logger.Fatal("Failed to insert data into geonames: %s", err)
				}
				logger.Debug(p.Sprintf("Inserted data for %d - %d", index-10000, index))
			}(data, i)
			wg.Add(1)
			data = []models.Geonames{}
		}
	}

	// Wait for all inserts to finish
	wg.Wait()
	if _, err := db.Struct.InsertSlice(&data).Run(); err != nil {
		logger.Fatal("Failed to insert data into geonames: %s", err)
	}

	logger.Info("Successfully inserted geodata")
}

func toInt(val string) int {
	rtc, err := strconv.Atoi(val)
	if err != nil {
		logger.Warning("Failed to convert %q to an integer: %s", val, err)
	}

	return rtc
}

func toFloat(val string) float64 {
	rtc, err := strconv.ParseFloat(val, 64)
	if err != nil {
		logger.Warning("Failed to convert %q to an float: %s", val, err)
	}

	return rtc
}
