package dbutils

import (
	"database/sql"

	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"git.rpjosh.de/RPJosh/workout/pkg/database/dbstruct"
)

type Db struct {
	*database.Utils

	Struct *dbstruct.Operator
}

// New initializes a new instance of the utils
func New(db *sql.DB) *Db {
	rtc := &Db{Utils: database.NewUtils(db)}
	rtc.Struct = dbstruct.NewOperator(rtc)

	return rtc
}

func NewByDb(db database.SqlConnection) *Db {
	rtc := &Db{Utils: database.NewUtilsByDb(db)}
	rtc.Struct = dbstruct.NewOperator(rtc)

	return rtc
}

func (d *Db) NewTransaction() (*Db, error) {
	newUtils, err := d.Utils.NewTransaction()

	rtc := &Db{Utils: newUtils}
	rtc.Struct = dbstruct.NewOperator(rtc)

	return rtc, err
}
