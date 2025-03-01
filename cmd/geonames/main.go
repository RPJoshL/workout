package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"git.rpjosh.de/RPJosh/go-ddl-parser"
	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api"
	"git.rpjosh.de/RPJosh/workout/internal/dbutils"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
	"github.com/guregu/null/v5"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const City1000 = "./dependencies/cities1000.txt"
const CountryAll = "./dependencies/allCountries.txt"

func main() {
	defer logger.CloseFile()

	// Get the generic configuration of the app
	conf := models.GetAppConfig()
	api := api.Api{Config: conf}
	db := dbutils.New(api.GetDb())

	// Data to insert
	data := []models.Geonames{}

	// Cached data we inserted into the database
	dataAll := []models.Geonames{}
	// Map with indexed adm4 codes to improve performance
	adm4Cache := map[string]string{}
	adm3Cache := map[string]string{}

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
		vv := r.Text()
		vals := strings.Split(vv, "\t")

		// Get administration types
		adm0Id, adm1Id, adm2Id, adm3Id, adm4Id := getAdministrationCodes(vals)

		d := models.Geonames{
			Geonameid:  utils.ToInt(vals[0]),
			Name:       vals[1],
			Location:   ddl.Location{Longitude: utils.ToFloat(vals[5]), Latitude: utils.ToFloat(vals[4])},
			Country:    vals[8],
			Population: utils.ToInt(vals[14]),
			Adm1:       null.StringFrom(adm1Id),
			Adm2:       null.StringFrom(adm2Id),
			Adm3:       null.StringFrom(adm3Id),
			Adm4:       null.StringFrom(adm4Id),
		}

		// Alternative names
		if vals[3] != "" {
			d.Alternatenames = null.StringFrom(vals[3])
		}

		data = append(data, d)
		dataAll = append(dataAll, d)

		// Add to caching map
		adm4Cache[fmt.Sprintf("%s#%s#%s#%s#%s", adm0Id, adm1Id, adm2Id, adm3Id, adm4Id)] = "YES"
		adm3Cache[fmt.Sprintf("%s#%s#%s#%s", adm0Id, adm1Id, adm2Id, adm3Id)] = "YES"

		// Insert 5.000 Rows at once. This should have the best performance
		if i%5000 == 0 {
			// Pass by copy
			wg.Add(1)
			go func(dd []models.Geonames, index int) {
				defer wg.Done()

				p := message.NewPrinter(language.English)
				if _, err := db.Struct.InsertSlice(&dd).Run(); err != nil {
					logger.Fatal("Failed to insert data into geonames: %s", err)
				}
				logger.Debug("%s", p.Sprintf("Inserted data for %d - %d", index-5000, index))
			}(data, i)
			data = []models.Geonames{}
		}

		if i%20000 == 0 {
			wg.Wait()
		}
	}

	// Wait for all inserts to finish
	wg.Wait()
	if _, err := db.Struct.InsertSlice(&data).Run(); err != nil {
		logger.Fatal("Failed to insert data into geonames: %s", err)
	}

	logger.Info("Successfully inserted geodata (cities)")

	// Check if we should proceed with parsing countries data (additional data)
	if os.Getenv("DISABLE_COUNTRIES") == "true" {
		logger.Debug("Disabled country parsing")
		return
	}

	// Truncate table
	if _, err := db.Db.GetDb().Exec("TRUNCATE TABLE geonames_adm"); err != nil {
		logger.Fatal("Failed to truncate table data for 'geonames_adm'")
	}

	// Read country file
	logger.Info("Reading country data...")
	fileCountry, err := os.Open(CountryAll)
	if err != nil {
		logger.Fatal("Failed to open country file: %s", err)
	}
	defer fileCountry.Close()

	// Data to insert
	dataCountry := []models.GeonamesAdm{}

	r = bufio.NewScanner(fileCountry)
	i = 0
	iFound := 0
	p := message.NewPrinter(language.English)
	for r.Scan() {
		i++
		// Seperated by tab
		str := r.Text()
		vals := strings.Split(str, "\t")

		// Get administration type
		typ := vals[7]
		adm0Id, adm1Id, adm2Id, adm3Id, adm4Id := getAdministrationCodes(vals)

		// Check if we need to store this type in the database
		found := false
		value := ""
		var adm2, adm3 null.String
		switch typ {
		case "ADM4":
			_, found = adm4Cache[fmt.Sprintf("%s#%s#%s#%s#%s", adm0Id, adm1Id, adm2Id, adm3Id, adm4Id)]
			value = adm4Id
			adm3 = null.StringFrom(adm3Id)
			adm2 = null.StringFrom(adm2Id)
		case "ADM3":
			_, found = adm3Cache[fmt.Sprintf("%s#%s#%s#%s", adm0Id, adm1Id, adm2Id, adm3Id)]
			value = adm3Id
			adm2 = null.StringFrom(adm2Id)
		case "ADM2":
			for _, d := range dataAll {
				if d.Adm2.String == adm2Id && d.Adm1.String == adm1Id && d.Country == adm0Id {
					found = true
					break
				}
			}
			value = adm2Id
		case "ADM1":
			for _, d := range dataAll {
				if d.Adm1.String == adm1Id && d.Country == adm0Id {
					found = true
					break
				}
			}
			value = adm1Id
		}
		if len(value) > 20 {
			logger.Error("Got to long value: %q - %q - %q!!", value, vals[0], typ)
			fmt.Println(str)
			return
		}

		// Insert data into database
		if found {
			d := models.GeonamesAdm{
				Geonameid: utils.ToInt(vals[0]),
				Typ:       typ,
				Value:     value,
				Name:      vals[1],
				Adm0:      adm0Id,
				Adm1:      adm1Id,
				Adm2:      adm2,
				Adm3:      adm3,
			}
			if vals[3] != "" {
				if len(vals[3]) > 4000 {
					// It could happen that we get an invalid string when splitting incorrect
					for i := 4000; i > 3900; i-- {
						vv := (vals[3])[:i]
						if utf8.ValidString(vv) {
							d.Alternatenames = null.StringFrom(vv)
							break
						}
					}
				} else {
					d.Alternatenames = null.StringFrom(vals[3])
				}
			}
			dataCountry = append(dataCountry, d)
			iFound++
		}

		// Insert 5.000 Rows at once. This should have the best performance
		if len(dataCountry) > 0 && iFound%5000 == 0 {
			// Pass by copy
			wg.Add(1)
			go func(dd []models.GeonamesAdm, index int) {
				defer wg.Done()

				p := message.NewPrinter(language.English)
				if _, err := db.Struct.InsertSlice(&dd).Run(); err != nil {
					logger.Fatal("Failed to insert data into geonames_adm: %s", err)
				}
				logger.Debug("%s", p.Sprintf("Inserted data for %d - %d", index-5000, index))
			}(dataCountry, iFound)
			dataCountry = []models.GeonamesAdm{}
		}

		// Print status every 1 million rows. We have ~12 millions
		if i%1000000 == 0 {
			logger.Debug("%s", p.Sprintf("Processed data for %d countries", i))
		}
	}
	wg.Wait()
	// Insert last ones
	if len(dataCountry) > 0 {
		if _, err := db.Struct.InsertSlice(&dataCountry).Run(); err != nil {
			logger.Fatal("Failed to insert data into geonames_adm: %s", err)
		}
	}

	logger.Info("Successfully inserted geonames (countries)")

}

// getAdministrationCodes returns the administration values for the provided line.
//
//   - ADM0: Land
//   - ADM1: Bundesland
//   - ADM2: Regierungsbezirk
//   - ADM3: Landkreis
//   - ADM4: Gemeinde
//
// A detailed description can be found here: https://www.geonames.org/export/codes.html
func getAdministrationCodes(vals []string) (adm0, adm1, adm2, adm3, adm4 string) {
	adm4 = vals[13]
	adm3 = vals[12]
	adm2 = vals[11]
	adm1 = vals[10]
	adm0 = vals[8]

	// Some rows do have an invalid adm2 code
	if len(adm2) > 15 {
		adm2 = ""
	}

	return
}
