package migrateregistry

import "github.com/gocaveman/caveman/migrate"

var global migrate.MigrationList

func MustRegister(m migrate.Migration) migrate.Migration {
	global = append(global, m)
	return m
}

func MustRegisterList(ml migrate.MigrationList) migrate.MigrationList {
	global = append(global, ml...)
	return ml
}

func Contents() migrate.MigrationList {
	return global.Sorted()
}
