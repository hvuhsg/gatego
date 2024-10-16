package gatego

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hvuhsg/gatego/pkg/cron"
)

type Check struct {
	Name    string
	Cron    string
	URL     string
	Method  string
	Timeout time.Duration
	Headers map[string]string
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

type Checker struct {
	Delay     time.Duration
	Checks    []Check
	scheduler *cron.Cron
	OnFailure func(error)
}

func (c Checker) Start() error {
	c.scheduler = cron.New()

	for _, check := range c.Checks {
		err := c.scheduler.Add(uuid.NewString(), check.Cron, check.run(c.OnFailure))
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
