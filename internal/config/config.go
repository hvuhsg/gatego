package config

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hvuhsg/gatego/internal/middlewares"
	"github.com/hvuhsg/gatego/pkg/cron"
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

type Check struct {
	Name      string            `yaml:"name"`
	Cron      string            `yaml:"cron"`
	URL       string            `yaml:"url"`
	Method    string            `yaml:"method"`
	Timeout   time.Duration     `yaml:"timeout"`
	Headers   map[string]string `yaml:"headers"`
	OnFailure string            `yaml:"on_failure"`
}

func (c Check) validate() error {
	if len(c.Name) == 0 {
		return errors.New("check requires a name")
	}

	if _, err := cron.NewSchedule(c.Cron); err != nil {
		return errors.New("invalid check cron expression")
	}

	if !isValidURL(c.URL) {
		return errors.New("invalid check url")
	}

	if !isValidMethod(c.Method) {
		return errors.New("invalid check method")
	}

	return nil
}

type Path struct {
	Path        string             `yaml:"path"`
	Destination *string            `yaml:"destination"` // The domain / url of the service server
	Directory   *string            `yaml:"directory"`   // path to dir you want to serve
	Backend     *Backend           `yaml:"backend"`     // List of servers to load balance between
	Headers     *map[string]string `yaml:"headers"`
	OmitHeaders []string           `yaml:"omit_headers"` // Omit specified headers
	Minify      []string           `yaml:"minify"`
	Gzip        *bool              `yaml:"gzip"`
	Timeout     time.Duration      `yaml:"timeout"`
	MaxSize     uint64             `yaml:"max_size"`
	OpenAPI     *string            `yaml:"openapi"`
	RateLimits  []string           `yaml:"ratelimits"`
	Checks      []Check            `yaml:"checks"` // Automated checks
	Cache       bool               `yaml:"cache"`  // Cache responses that has cache headers
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

		if p.Cache {
			log.Println("[WARNING] Using cache while serving static files is not recommanded")
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

	for _, check := range p.Checks {
		if err := check.validate(); err != nil {
			return err
		}
	}

	return nil
}

type AnomalyDetection struct {
	HeaderName        string `yaml:"header_name"`
	MinScore          int    `yaml:"min_score"`
	MaxScore          int    `yaml:"max_score"`
	TresholdForRating int    `yaml:"treshold_for_rating"`
	Active            bool   `yaml:"active"`
}

func (a *AnomalyDetection) validate() error {
	if a.HeaderName == "" {
		a.HeaderName = "X-Anomaly-Score"
	}

	if a.MinScore == 0 {
		a.MinScore = 100
	}

	if a.MaxScore == 0 {
		a.MaxScore = 200
	}

	if a.TresholdForRating == 0 {
		a.TresholdForRating = 100
	}

	if a.MaxScore <= a.MinScore {
		return errors.New("anomaly detection maxScore MUST be grater the minScore")
	}

	return nil
}

type Service struct {
	Domain           string            `yaml:"domain"` // The domain / host the request was sent to
	Paths            []Path            `yaml:"endpoints"`
	AnomalyDetection *AnomalyDetection `yaml:"anomaly_detection"`
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

	if s.AnomalyDetection != nil {
		if err := s.AnomalyDetection.validate(); err != nil {
			return err
		}
	}

	return nil
}

type TLS struct {
	Auto     bool     `yaml:"auto"`
	Domains  []string `yaml:"domain"`
	Email    *string  `yaml:"email"`
	KeyFile  *string  `yaml:"keyfile"`
	CertFile *string  `yaml:"certfile"`
}

func (tls TLS) validate() error {
	if tls.Auto {
		if len(tls.Domains) == 0 {
			return errors.New("when using the auto tls feature you MUST include a list of domains to issue certificates for")
		}
		if tls.Email == nil || len(*tls.Email) == 0 || !isValidEmail(*tls.Email) {
			return errors.New("when using the auto tls feature you MUST include a valid email for the lets-encrypt registration")
		}
	}

	if tls.CertFile != nil {
		if tls.KeyFile == nil {
			return errors.New("you MUST provide certfile AND keyfile")
		}
	}

	if tls.KeyFile != nil {
		if tls.CertFile == nil {
			return errors.New("you MUST provide certfile AND keyfile")
		}

		if !isValidFile(*tls.CertFile) {
			return errors.New("certfile path is invalid")
		}

		if !isValidFile(*tls.KeyFile) {
			return errors.New("keyfile path is invalid")
		}
	}

	return nil
}

type OTEL struct {
	Endpoint    string  `yaml:"endpoint"`
	SampleRatio float64 `yaml:"sample_ratio"`
}

func (otel OTEL) validate() error {
	if len(otel.Endpoint) > 0 {
		if err := isValidGRPCAddress(otel.Endpoint); err != nil {
			return err
		}
	}

	if otel.SampleRatio < 0 {
		return errors.New("OpenTelemetry sample ratio MUST be above 0")
	}

	if otel.SampleRatio == 0 {
		return errors.New("OpenTelemetry sample ratio is missing or equales to 0")
	}

	if otel.SampleRatio > 1 {
		return errors.New("OpenTelemetry sample ratio CAN NOT be above 1")
	}

	return nil
}

type Config struct {
	Version string `yaml:"version"`
	Host    string `yaml:"host"` // listen host
	Port    uint16 `yaml:"port"` // listen port

	OTEL *OTEL `yaml:"open_telemetry"`

	// TLS options
	TLS TLS `yaml:"ssl"`

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

	if c.OTEL != nil {
		if err := (*c.OTEL).validate(); err != nil {
			return err
		}
	}

	if c.Port == 0 {
		return errors.New("port is required")
	}

	if err := c.TLS.validate(); err != nil {
		return err
	}

	if c.TLS.Auto && c.Port != 443 {
		return errors.New("the auto tls feature is only available if the server runs on port 443")
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

func isValidMethod(method string) bool {
	methods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}

	return slices.Contains(methods, method)
}

func isValidGRPCAddress(address string) error {
	if address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	// Split host and port
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid address format: %v", err)
	}

	// Validate port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port number: %v", err)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port number must be between 1 and 65535")
	}

	// Empty host means localhost/0.0.0.0, which is valid
	if host == "" {
		return nil
	}

	// Check if host is IPv4 or IPv6
	if ip := net.ParseIP(host); ip != nil {
		return nil
	}

	// Validate hostname format
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-\.]*[a-zA-Z0-9])?$`)
	if !hostnameRegex.MatchString(host) {
		return fmt.Errorf("invalid hostname format")
	}

	// Check hostname length
	if len(host) > 253 {
		return fmt.Errorf("hostname too long")
	}

	// Validate hostname parts
	parts := strings.Split(host, ".")
	for _, part := range parts {
		if len(part) > 63 {
			return fmt.Errorf("hostname label too long")
		}
	}

	return nil
}

func isValidEmail(email string) bool {
	// Define a regular expression for valid email addresses
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Match the email string with the regular expression
	return emailRegex.MatchString(email)
}
