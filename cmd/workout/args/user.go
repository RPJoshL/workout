package args

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"git.rpjosh.de/RPJosh/go-ddl-parser"
	"git.rpjosh.de/RPJosh/go-ddl-parser/structt"
	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/user"
	"git.rpjosh.de/RPJosh/workout/internal/models"
)

// User contains entry options for the CLI
type User struct {
	UserCreate UserCreate `cli:"create,c"`
}

type UserCreate struct {
}

func (e *UserCreate) SetUserCreate(cli *Cli) string {
	// Api connection
	userApi := user.Api{}
	cli.InjectApi(&userApi)

	// Get all fields to create the user and ask the user about them
	user := &models.User{}
	userRef := reflect.ValueOf(user)
	userType := userRef.Type().Elem()

	// Get database comment to display as a help
	mdb := ddl.NewMariaDb(userApi.R().Db.Db.GetDb())
	mField, _ := userType.FieldByName(structt.MetadataFieldName)
	metadata := structt.FromMetadataTag(mField.Tag.Get(structt.MetadataTagId))
	dbCols, err := mdb.GetTable(metadata.Schema, metadata.Table)
	if err != nil {
		logger.Fatal("Failed to get columns: %s", err)
	}

	// Loop over all fields and ask user
	reader := bufio.NewReader(os.Stdin)
	for i := range userType.NumField() {
		field := userType.Field(i)

		// Get tag
		valTag, ok := field.Tag.Lookup(structt.ColumnTagId)
		if !ok {
			continue
		}

		// Get column
		col := structt.FromColumnTag(valTag)
		if col.AutoIncrement {
			continue
		}

		// Get any comment
		for _, dbCol := range dbCols.Columns {
			if dbCol.Name == col.Name && dbCol.Comment != "" {
				for _, colLine := range strings.Split(dbCol.Comment, "\n") {
					fmt.Println("// " + colLine)
				}
			}
		}

		// Scan value
		fmt.Printf("%-12s ", field.Name+":")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		userRef.Elem().Field(i).Set(reflect.ValueOf(convertToVal(text, field.Type)))
		fmt.Println()
	}

	if err := userApi.CreateUser(*user); err != nil {
		logger.Fatal("Failed to create user: %s", err)
		os.Exit(1)
	}
	os.Exit(0)
	return ""
}

func convertToVal(val string, typ reflect.Type) any {
	var rv = reflect.New(typ)

	switch rv.Elem().Interface().(type) {
	case string:
		return val
	case int:
		if intVal, err := strconv.Atoi(val); err != nil {
			logger.Fatal("Failed to parse integer value %q: %s", val, err)
		} else {
			return intVal
		}
	}

	logger.Warning("Unknown type %s", typ)
	return nil
}
