package migrations

import (
	"fmt"
)

var Migrations = map[string]migration{}

type migration struct {
	UpSQL   string
	DownSQL string
}

func appendMigration(name, upSQL, downSQL string) {
	_, ok := Migrations[name]
	if ok {
		panic(fmt.Sprintf("cannot append migration: name '%s' already exists", name))
	}
	Migrations[name] = migration{
		UpSQL:   upSQL,
		DownSQL: downSQL,
	}
}
