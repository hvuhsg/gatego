# Cron

A Go package that implements a crontab-like service to execute and schedule repetitive tasks/jobs.

## Features

- Supports cron expressions for flexible scheduling
- Allows registering and managing multiple jobs
- Provides macros for common schedule patterns
- Supports custom timezones
- Allows setting custom tick intervals
- Supports starting and stopping the cron service

## Installation

```sh
go get github.com/hvuhsg/gatego/pkg/cron
```

## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/hvuhsg/gatego/pkg/cron"
)

func main() {
	c := cron.New()

	// Register a job
	c.MustAdd("job1", "*/5 * * * *", func() {
		fmt.Println("Running job1...")
	})

	// Set a custom timezone
	loc, _ := time.LoadLocation("Asia/Tokyo")
	c.SetTimezone(loc)

	// Set a custom tick interval
	c.SetInterval(5 * time.Second)

	// Start the cron service
	c.Start()

	// Stop the cron service after 30 seconds
	time.Sleep(30 * time.Second)
	c.Stop()
}
```

## Cron Expression Format

The package supports the following cron expression format:

```
* * * * *
│ │ │ │ │
│ │ │ │ └── Day of Week (0-6)
│ │ │ └──── Month (1-12)
│ │ └────── Day of Month (1-31)
│ └──────── Hour (0-23)
└────────── Minute (0-59)
```

It also supports the following macros:

- `@yearly` or `@annually`: Run once a year at midnight on the first day of the year
- `@monthly`: Run once a month at midnight on the first day of the month
- `@weekly`: Run once a week at midnight on Sunday
- `@daily` or `@midnight`: Run once a day at midnight
- `@hourly`: Run once an hour at the beginning of the hour
- `@minutely`: Run once a minute at the beginning of the minute
