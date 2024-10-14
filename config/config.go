package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hvuhsg/gatego/middlewares"
	"gopkg.in/yaml.v3"
)

const DefaultTimeout = time.Second * 30
const DefaultMaxRequestSize = 1024 * 10 // 10 MB
var SupportedBalancePolicies = []string{"round-robin", "random", "least-latency"}

type Backend struct {
	BalancePolicy string `yaml:"balance_policy"`
	Servers       []struct {
		URL    string `yaml:"url"`
		Weight uint   `yaml:"weight"`
	}
}

func (b Backend) validate() error {
	if !slices.Contains(SupportedBalancePolicies, b.BalancePolicy) {
		return fmt.Errorf("balance policy '%s' is not supported", b.BalancePolicy)
	}

	if len(b.Servers) == 0 {
		return errors.New("backend require at least one server")
	}

	for _, server := range b.Servers {
		if !isValidURL(server.URL) {
			return fmt.Errorf("invalid backend server url '%s'", server.URL)
		}
	}

	return nil
}

type Path struct {
	Path        string             `yaml:"path"`
	Destination *string            `yaml:"destination"` // The domain / url of the service server
	Directory   *string            `yaml:"directory"`   // path to dir you want to serve
	Backend     *Backend           `yaml:"backend"`     // List of servers to load balance between
	Headers     *map[string]string `yaml:"headers"`
	Minify      []string           `yaml:"minify"`
	Gzip        *bool              `yaml:"gzip"`
	Timeout     time.Duration      `yaml:"timeout"`
	MaxSize     uint64             `yaml:"max_size"`
	OpenAPI     *string            `yaml:"openapi"`
	RateLimits  []string           `yaml:"ratelimits"`
}

func (p Path) validate() error {
	if p.Path[0] != '/' {
		return errors.New("path must start with '/'")
	}

	if p.Destination != nil {
		if !isValidURL(*p.Destination) {
			return errors.New("invalid destination url")
		}

		if p.Directory != nil {
			return errors.New("can't have destination and directory for the same path")
		}
	}

	if p.Directory != nil {
		if !isValidDir(*p.Directory) {
			return errors.New("invalid directory path")
		}
	}

	if p.Backend != nil {
		if err := p.Backend.validate(); err != nil {
			return err
		}
	}

	if p.Destination == nil && p.Directory == nil && p.Backend == nil {
		return errors.New("path must have destination or directory or backend")
	}

	if p.OpenAPI != nil {
		if *p.OpenAPI == "" {
			return errors.New("openapi can't be empty (remove or fill)")
		}

		if !isValidFile(*p.OpenAPI) {
			return errors.New("invalid openapi spec path")
		}
	}

	for _, ratelimit := range p.RateLimits {
		_, err := middlewares.ParseLimitConfig(ratelimit)
		if err != nil {
			return fmt.Errorf("invalid ratelimit: %s", err.Error())
		}
	}

	return nil
}

type Service struct {
	Domain string `yaml:"domain"` // The domain / host the request was sent to
	Paths  []Path `yaml:"endpoints"`
}

func (s Service) validate() error {
	if !isValidHostname(s.Domain) {
		return errors.New("invalid domain")
	}

	for _, path := range s.Paths {
		if err := path.validate(); err != nil {
			return err
		}
	}

	return nil
}

type TLS struct {
	KeyFile  *string `yaml:"keyfile"`
	CertFile *string `yaml:"certfile"`
}

type Config struct {
	Version string `yaml:"version"`
	Host    string `yaml:"host"` // listen host
	Port    uint16 `yaml:"port"` // listen port

	// TLS options
	SSL TLS `yaml:"ssl"`

	Services []Service `yaml:"services"`
}

func (c Config) Validate(currentVersion string) error {
	if c.Version == "" {
		return errors.New("version is required")
	}

	progVersion, _ := version.NewVersion(currentVersion)
	configVersion, err := version.NewVersion(c.Version)
	if err != nil {
		return errors.New("version is invalid")
	}

	if configVersion.Compare(progVersion) > 0 {
		return errors.New("config version is not supported (too advanced)")
	}

	if c.Host == "" {
		return errors.New("host is required")
	}

	for _, service := range c.Services {
		if err := service.validate(); err != nil {
			return err
		}
	}

	return nil
}

func ParseConfig(filepath string, currentVersion string) (Config, error) {
	// Read the YAML file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return Config{}, err
	}

	// Defaults
	c := Config{Port: 80}

	// Unmarshal the YAML data into the struct
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return Config{}, err
	}

	if err := c.Validate(currentVersion); err != nil {
		return Config{}, err
	}

	return c, nil
}

func isValidHostname(hostname string) bool {
	// Remove leading/trailing whitespace
	hostname = strings.TrimSpace(hostname)

	// Check if the hostname is empty
	if hostname == "" {
		return false
	}

	// Check if the hostname is too long (max 253 characters)
	if len(hostname) > 253 {
		return false
	}

	// Check for localhost
	if hostname == "localhost" {
		return true
	}

	// Check if it's an IP address (IPv4 or IPv6)
	if ip := net.ParseIP(hostname); ip != nil {
		return true
	}

	// Regular expression for domain validation
	// This regex allows for domains with multiple subdomains and supports IDNs
	domainRegex := regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$`)

	return domainRegex.MatchString(hostname)
}

func isValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isValidDir(path string) bool {
	if path == "" {
		return false
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func isValidFile(path string) bool {
	if path == "" {
		return false
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !fileInfo.IsDir()
}
