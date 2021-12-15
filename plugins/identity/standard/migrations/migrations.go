package migrations

import (
	"fmt"
	"runtime"
)

var Migrations = map[string]migration{}

type migration struct {
	UpSQL   string
	DownSQL string
}

func appendMigration(upSQL, downSQL string) {
	_, filename, _, ok := runtime.Caller(1)
	if ok {
		_, ok := Migrations[filename]
		if ok {
			panic(fmt.Sprintf("cannot append migration: name '%s' already exists", filename))
		}
	}

	Migrations[filename] = migration{
		UpSQL:   upSQL,
		DownSQL: downSQL,
	}

}
