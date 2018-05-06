package gen

import (
	"flag"
	"path/filepath"
)

func init() {
	globalMapGenerator["main-rest-sqlite3"] = GeneratorFunc(func(s *Settings, name string, args ...string) error {

		// FIXME: see if we want to move to pflag - it's probably better and
		// would be smart to be consistent
		fset := flag.NewFlagSet("gen", flag.ContinueOnError)
		// storeName := fset.String("name", "Store", "The name of the store struct to define")
		targetFile, data, err := ParseFlagsAndOneFile(s, fset, args)
		if err != nil {
			return err
		}

		targetDir, _ := filepath.Split(targetFile)
		_, targetDirName := filepath.Split(targetDir)

		data["TargetDirName"] = targetDirName

		return OutputGoSrcTemplate(s, data, targetFile, `
package main

import (
	"github.com/gocraft/dbr"
	"github.com/gocaveman/dbrobj"

	_ "github.com/mattn/go-sqlite3"
	// _ "github.com/go-sql-driver/mysql"
	// _ "github.com/lib/pq"
)

const APP_NAME = "{{.TargetDirName}}"

func main() {

	var err error

	// TODO: HTTPS support, including certificates - letsencrypt with option to override
	// and workable solution for local development
	pflag.StringP("http-listen", "l", ":8080", "IP:Port to listen on for HTTP")
	// pflag.StringP("db-dsn", "", "root:@tcp(localhost:3306)/"+APP_NAME+"?charset=utf8mb4,utf8", "Database connection string")
	pflag.StringP("db-dsn", "", "file:"+APP_NAME+"?mode=memory&cache=shared", "Database connection string")
	pflag.StringP("db-driver", "", "sqlite3", "Database driver name")
	// TODO: on the migrations we probably also want a way to separately check, and then update
	// without running the app - so you can build the new version with the schema changes, check it
	// apply it, and then deploy.
	pflag.StringP("db-migrate", "", "auto", "Database migration behavior ('auto' to update, 'check' to report out of date, or 'none' to ignore migrations)")
	pflag.BoolP("debug", "g", false, "Enable debug output (intended for development only)")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	// TODO: add viper config file search and load stuff

	if wd := viper.GetString("workdir"); wd != "" {
		if err := os.Chdir(wd); err != nil {
			log.Fatalf("Error changing directory to %q: %v", wd, err)
		}
	}

	hl := webutil.NewDefaultHandlerList()
	for _, item := range handlerregistry.Contents() {
		hl = append(hl, item.Value)
	}
	hl = append(hl, http.NotFoundHandler())

	dbDriver := viper.GetString("db-driver")
	dbDsn := viper.GetString("db-dsn")
	// TODO: provide in autowire, just need to know the pattern for names...

	dbrConn, err := dbr.Open(dbDriver, dbDsn, nil)
	if err != nil {
		log.Fatal(err)
	}
	autowire.Provide("", dbrConn)

	connector := dbrobj.NewConfig().NewConnector(dbrConn, nil)
	autowire.Provide("", connector)

	err = autowire.Contents().Run()
	if err != nil {
		log.Fatalf("autowire error: %v", err)
	}

	dbMigrateMode := viper.GetString("db-migrate")
	ml := migrateregistry.Contents().WithDriverName(dbDriver).Sorted()
	// EDITME: you can filter migrations here if needed
	versioner, err := migratedbr.New(dbDriver, dbDsn)
	if err != nil {
		log.Fatal(err)
	}
	runner := migrate.NewRunner(dbDriver, dbDsn, versioner, ml)
	if dbMigrateMode == "check" {
		result, err := runner.CheckAll(true)
		if err != nil {
			log.Fatalf("Migration check error: %v", err)
		}
		log.Printf("Migration check result: %+v", result) // TODO: better output
	} else if dbMigrateMode == "auto" {
		err := runner.RunAllUpToLatest()
		if err != nil {
			log.Fatalf("Migration auto-update error: %v", err)
		}
	}

	var wg sync.WaitGroup

	// FIXME: HTTP/HTTPS graceful shutdown;
	// example here https://gist.github.com/peterhellberg/38117e546c217960747aacf689af3dc2
	// The idea is that if we stop receiving new requests and succesfully complete the old,
	// a load balancer can reliably hold onto and retry connections that are rejected
	// and we can get a deploy/restart (even on a single server) with no actual "down time".
	// FIXME: we should also look at runtime.SetFinalizer and anything else relevant and
	// see if there is some sort of "shutdown" hook that can easily be put together - without
	// getting too weird and complicated - maybe it's just a matter of calling Close() on
	// whatever needs it as defer in main and things get cleaned up that way (and possibly
	// autowire needs to support some sort of Close() mechanism as well)
	webutil.StartHTTPServer(&http.Server{
		Addr:    viper.GetString("http-listen"),
		Handler: hl,
	}, &wg)

	wg.Wait()
}

`, false)

	})
}
