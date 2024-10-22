package gatego

import (
	"github.com/hvuhsg/gatego/config"
	"github.com/hvuhsg/gatego/pkg/monitor"
)

func createMonitorChecks(services []config.Service) []monitor.Check {
	checks := make([]monitor.Check, 0)
	for _, service := range services {
		for _, path := range service.Paths {
			for _, checkConfig := range path.Checks {
				check := monitor.Check{
					Name:      checkConfig.Name,
					Cron:      checkConfig.Cron,
					URL:       checkConfig.URL,
					Method:    checkConfig.Method,
					Timeout:   checkConfig.Timeout,
					Headers:   checkConfig.Headers,
					OnFailure: checkConfig.OnFailure,
				}

				checks = append(checks, check)
			}
		}
	}

	return checks
}
