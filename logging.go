// +build !windows

package caveman

import (
	"log"
	"log/syslog"
)

func LogToSyslog(prgName string) error {
	logwriter, err := syslog.New(syslog.LOG_NOTICE, prgName)
	if err != nil {
		return err
	}
	log.SetOutput(logwriter)
	return nil
}
