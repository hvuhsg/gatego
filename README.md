# Reverse Proxy Server

## Overview

This reverse proxy server is designed to forward incoming requests to internal services, while offering advanced features such as SSL termination, rate limiting, content optimization, and OpenAPI-based request/response validation.

## Key Features
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
  minify:
    js: true
    html: true
    css: true
    json: true
    xml: true
    svg: true
    # You can use 'all: true' instaed to enable all content-types
  
  # Optional
  gzip: true  # Enable GZIP compression
```


### 3. Request Limits and Timeouts

- Timeout: Custom timeouts can be set to avoid slow backend services from hanging client requests.
- Maximum Request Size: Limits can be placed on the size of incoming requests to prevent excessively large payloads from overwhelming the server.

```yaml
- path: /
  timeout: 5s  # Custom timeout for backend responses (Default 30s)
  max-size: 2048  # Max request size in bytes (Default 10MB)
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

## Configuration Example

Here’s a generic example of how you can configure the reverse proxy:

```yaml
version: 'your-version'
host: your-host
port: your-port

ssl:
  keyfile: /path/to/your/ssl/keyfile
  certfile: /path/to/your/ssl/certfile

services:
  - domain: your-domain.com
    endpoints:
      - path: /your-endpoint  # will be served for every request with path that start with /your-endpoint (Example: /your-endpoint/1)
        destination: http://your-backend-service/
        
        minify:
          js: true
          html: true
          css: true
          json: true
          xml: true
          svg: true
          # You can use 'all: true' instaed to enable all content-types

        gzip: true  # Enable GZIP compression
        
        timeout: 5s  # Custom timeout for backend responses (Default 30s)
        max-size: 2048  # Max request size in bytes (Default 10MB)
        
        ratelimits:
          - ip-10/m  # Limit to 10 requests per minute per IP
          - ip-500/d  # Limit to 500 requests per day per IP
        
        openapi: /path/to/openapi.yaml  # OpenAPI file for request/response validation
```

### Breakdown:

- Services: You can define multiple domains and endpoints, each with their own routing and optimization settings.
- Endpoints: You can have multiple endpoints that share a path prefix, the request will be routed to the longest muching endpoint.

## License

This project is licensed under the MIT License.
