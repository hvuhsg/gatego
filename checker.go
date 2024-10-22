package gatego

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hvuhsg/gatego/pkg/cron"
)

type Check struct {
	Name      string
	Cron      string
	URL       string
	Method    string
	Timeout   time.Duration
	Headers   map[string]string
	OnFailure string
}

func (c Check) run(onFailure func(error)) func() {
	return func() {
		// Create a client with timeout
		client := &http.Client{
			Timeout: c.Timeout,
		}

		// Create new request
		req, err := http.NewRequest(c.Method, c.URL, nil)
		if err != nil {
			log.Default().Printf("Check <%s> error creating check request URL=%s Method=%s\n", c.Name, c.URL, c.Method)
			onFailure(err)
			return
		}

		// Add headers
		for key, value := range c.Headers {
			req.Header.Add(key, value)
		}

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			log.Default().Printf("Check <%s> error sending request Error=%s\n", c.Name, err.Error())
			onFailure(err)
			return
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			log.Default().Printf("Check <%s> failed. Expected status code 200 got %d\n", c.Name, resp.StatusCode)
			onFailure(fmt.Errorf("expected status code 200 got %d", resp.StatusCode))
			return
		}
	}
}

func handleFailure(check Check, err error) error {
	// Expand command
	command := check.OnFailure
	date := time.Now().UTC().Format("2006-01-02 15:04:05")
	command = strings.ReplaceAll(command, "$date", date)
	command = strings.ReplaceAll(command, "$error", err.Error())
	command = strings.ReplaceAll(command, "$check_name", check.Name)

	// Run it
	args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:]...)
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

type Checker struct {
	Delay     time.Duration
	Checks    []Check
	scheduler *cron.Cron
}

func (c Checker) Start() error {
	c.scheduler = cron.New()

	for _, check := range c.Checks {
		err := c.scheduler.Add(uuid.NewString(), check.Cron, check.run(func(err error) {
			if check.OnFailure != "" {
				if err := handleFailure(check, err); err != nil {
					log.Default().Printf("Failed to spawn on_failure command: %s\n", err)
				}
			}
		}))
		if err != nil {
			return err
		}
	}

	go func() {
		time.Sleep(c.Delay)
		c.scheduler.Start()
		log.Default().Println("Started running automated checks.")
	}()

	return nil
}
