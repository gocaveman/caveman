package migrateregistry

import "github.com/gocaveman/caveman/migrate"

var global migrate.MigrationList

func MustRegister(m migrate.Migration) {
	global = append(global, m)
}

func MustRegisterList(ml migrate.MigrationList) {
	global = append(global, ml...)
}

func Contents() migrate.MigrationList {
	return global.Sorted()
}
