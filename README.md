# Reverse Proxy Server

[![Tests](https://github.com/hvuhsg/gatego/actions/workflows/go-tests.yml/badge.svg?branch=main)](https://github.com/hvuhsg/gatego/actions/workflows/go-tests.yml)

## Overview

This reverse proxy server is designed to forward incoming requests to internal services, while offering advanced features such as SSL termination, rate limiting, content optimization, and OpenAPI-based request/response validation.

## Supported Features

- 🔒 SSL Termination - HTTPS support with configurable SSL certificates

- 🚀 Content Optimization
  - Minification for HTML, CSS, JS, XML, JSON, and SVG
  - GZIP compression support


- ⚡ Performance Controls
  - Configurable request timeouts
  - Maximum request size limits
  - Response caching for cacheable content


- 🛡️ Security & Protection

  - IP-based rate limiting (per minute/day)
  - Request/response validation via OpenAPI
  - Anomaly detection score (per session)

- ⚖️ Load Balancing

  - Multiple backend server support
  - Round-robin, random, and least-latency policies
  - Weighted distribution options


- 📁 File Serving - Static file serving with path stripping

- 🏥 Health Monitoring

  - Automated health checks with cron scheduling
Configurable failure notifications


- 📊 Observability - OpenTelemetry integration for tracing and metrics

## More About The Features
### 1. SSL Termination

The proxy supports secure connections through SSL, with configurable paths to the SSL key and certificate files. This allows for secure HTTPS communication between clients and the reverse proxy.

```yaml
# Optional
ssl:
  keyfile: /path/to/your/ssl/keyfile
  certfile: /path/to/your/ssl/certfile
```

### 2. Content Optimization

- Minification: The server can minify content (e.g., HTML, CSS, JavaScript, XML, JSON, SVG) before forwarding it to the client, reducing response sizes and improving load times.
- Compression: GZIP compression is supported to further reduce the size of responses, optimizing bandwidth usage.

```yaml
- path: /

  # Optional
  minify: [js, html, css, json, xml, svg]
  # You can use 'all' instaed to enable all content-types
  
  # Optional
  gzip: true  # Enable GZIP compression
```


### 3. Request Limits and Timeouts

- Timeout: Custom timeouts can be set to avoid slow backend services from hanging client requests.
- Maximum Request Size: Limits can be placed on the size of incoming requests to prevent excessively large payloads from overwhelming the server.

```yaml
- path: /
  timeout: 5s  # Custom timeout for backend responses (Default 30s)
  max_size: 2048  # Max request size in bytes (Default 10MB)
```

### 4. Rate Limiting

Rate limiting can be applied to prevent abuse, restricting the number of requests an individual client (based on IP) can make within a specific time window. Multiple rate limit policies can be configured, such as:
- Requests per minute from the same IP
- Requests per day from the same IP

```yaml
- path: /

  # Optional
  ratelimits:
    - ip-10/m  # Limit to 10 requests per minute per IP
    - ip-500/d  # Limit to 500 requests per day per IP
```

### 5. OpenAPI-based Request and Response Validation

The server integrates OpenAPI for validating incoming requests and outgoing responses against an OpenAPI specification document. This ensures that:

- Requests conform to the expected format, including parameters, headers, and body content.
- Responses adhere to the defined API schema, ensuring consistent and reliable data exchange.

You can specify the OpenAPI file path in the configuration, and the server will use it to validate the requests and responses automatically.

```yaml
- path: /

  # Optional
  openapi: /path/to/openapi.yaml  # OpenAPI file for request/response validation
```


### 6. Routing Anomaly Detection

The Server will calculate an anomaly score for the request based on global avg routing and session avg routing.
The score is added as a header to the request `X-Anomaly-Score`.
The score ranging between 0 (normal request) to 1 (a-normal request)

```yaml
services:
  - domain: your-domain.com
  
    # Will add to downstream request an header with routing anomaly score between 0 (normal) and 1 (suspicuse)
    anomaly_detection: 
      active: true
      header_name: "X-Anomaly-Score" # (Optional) [Default: X-Anomaly-Score]
      min_score: 100 # (Optional) Every internal score below this number is 0 [Default: 100]
      max_score: 100 # (Optional) Every internal score above this number is 1 [Default: 200]
      treshold_for_rating: 100 # (Optional) The amount of requests to collect stats on before starting to rate anomaly [Default: 100]
```


### 7. Load Balancing and File Serving

File serving is used when the `directory` field is set.
> The endpoint path is removed from the request path before the file lookup. For example a path of /static and request path of /static/file.txt and a directory /var/www will search the file in /var/www/file.txt and not /var/www/static/file.txt

```yaml
- path: /static
  directory: /var/www/
```

The Server support load balancing between a number of backend servers and allow you to choose the balancing policy.


```yaml
- path: /static
  backend:
    balance_policy: 'round-robin'
    servers:
      - url: http://backend-server-1/
        weight: 1
      - url: http://backend-server-2/
        weight: 2
```

#### Supported Policies:
- `round-robin` (affected by weights)
- `random` (affected by weights)
- `least-latency` (**not** affected by weights)


### 8. Health Checks

The server supports automated health checks for backend services. You can configure periodic checks to monitor the health of your backend servers under each endpoint's configuration.

```yaml
- path: /
  checks:
    - name: "Health Check"      # Descriptive name for the check
      cron: "* * * * *"        # Cron expression for check frequency
      # Supported cron macros:
      # - @yearly (or @annually) - Run once a year
      # - @monthly              - Run once a month
      # - @weekly               - Run once a week
      # - @daily                - Run once a day
      # - @hourly               - Run once an hour
      # - @minutely             - Run once a minute
      method: GET              # HTTP method for the health check
      url: "http://backend-server-1/up"  # Health check endpoint
      timeout: 5s             # Timeout for health check requests
      headers:                # Optional custom headers
        Host: domain.org
        Authorization: "Bearer abc123"
```

### 9. OpenTelemetry Integration
The server includes built-in support for OpenTelemetry, enabling comprehensive observability through distributed tracing, metrics, and logging. This integration helps monitor application performance, troubleshoot issues, and understand system behavior in distributed environments.

```yaml
version: '...'

open_telemetry:
  endpoint: "localhost:4317"
  sample_ratio: 0.01  # == 1%
```

## Configuration Example

Here’s a generic example of how you can configure the reverse proxy:

```yaml
version: '0.0.1'
host: your-host
port: your-port

ssl:
  keyfile: /path/to/your/ssl/keyfile
  certfile: /path/to/your/ssl/certfile

open_telemetry:
  endpoint: "localhost:4317"
  sample_ratio: 0.01  # == 1%

services:
  - domain: your-domain.com

    # Will add to downstream request an header with routing anomaly score between 0 (normal) and 1 (suspicuse)
    anomaly_detection: 
      active: true
      header_name: "X-Anomaly-Score" # (Optional) [Default: X-Anomaly-Score]
      min_score: 100 # (Optional) Every internal score below this number is 0 [Default: 100]
      max_score: 100 # (Optional) Every internal score above this number is 1 [Default: 200]
      treshold_for_rating: 100 # (Optional) The amount of requests to collect stats on before starting to rate anomaly [Default: 100]
  
    endpoints:
      - path: /your-endpoint  # will be served for every request with path that start with /your-endpoint (Example: /your-endpoint/1)

        # directory: /home/yoyo/  # For static files serving
        # destination: http://your-backend-service/
        backend:
          balance_policy: 'round-robin'  # Can be 'round-robin', 'random', or 'least-latency'
          servers:
            - url: http://backend-server-1/
              weight: 1
            - url: http://backend-server-2/
              weight: 2
        
        minify: [js, html, css, json, xml, svg]
        # You can use 'all' instaed to enable all content-types

        gzip: true  # Enable GZIP compression
        
        timeout: 5s  # Custom timeout for backend responses (Default 30s)
        max_size: 2048  # Max request size in bytes (Default 10MB)
        
        ratelimits:
          - ip-10/m  # Limit to 10 requests per minute per IP
          - ip-500/d  # Limit to 500 requests per day per IP
        
        openapi: /path/to/openapi.yaml  # OpenAPI file for request/response validation

        omit_headers: [Server]  # Omit response headers

        checks:
          - name: "Health Check"
            
            cron: "* * * * *" # == @minutely
            # Support cron format and macros.
            # Macros:
            # - @yearly
            # - @annually
            # - @monthly
            # - @weekly
            # - @daily
            # - @hourly
            # - @minutely

            method: GET  # HTTP Method
            url: "http://backend-server-1/up"
            timeout: 5s
            headers:
              Host: domain.org
              Authorization: "Bearer abc123"

            # on_failure runs a shell command if the check fails. Expands $date, $error, $check_name.
            on_failure: |
              curl -d "Health check '$check_name' failed at $date due to: $error" ntfy.sh/gatego
        cache: true  # Cache responses that has cache headers (Cache-Control and Expire)

```

### Breakdown
The configuration is organized into three main sections:

- Global Settings:
  - Server configuration (host, port) 
  - SSL settings 
  - OpenTelemetry configuration


- Services
  - Domain-based routing
  - Multiple endpoints per domain
  - Path-based matching with longest-prefix wins


- Endpoints
  - Backend service configuration
  - Performance optimizations
  - Security controls
  - Monitoring settings

Each endpoint can be independently configured with its own set of features, allowing for flexible and granular control over different parts of your application.

## License

This project is licensed under the MIT License.
