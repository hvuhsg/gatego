package cron

const (
	Yearly   = "@yearly"
	Annually = "@annually"
	Monthly  = "@monthly"
	Weekly   = "@weekly"
	Daily    = "@daily"
	Midnight = "@midnight"
	Hourly   = "@hourly"
	Minutely = "@minutely"
)

var macros = map[string]string{
	Yearly:   "0 0 1 1 *",
	Annually: "0 0 1 1 *",
	Monthly:  "0 0 1 * *",
	Weekly:   "0 0 * * 0",
	Daily:    "0 0 * * *",
	Midnight: "0 0 * * *",
	Hourly:   "0 * * * *",
	Minutely: "* * * * *",
}
