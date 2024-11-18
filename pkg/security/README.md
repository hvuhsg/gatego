
# Security Package

This Go package provides various security-related functions for hashing, message authentication codes (MAC), JSON Web Tokens (JWT), and random string generation.

## Functions

### S256Challenge

```go
func S256Challenge(code string) string
```

`S256Challenge` creates a base64-encoded SHA-256 challenge string derived from the provided code. The padding of the resulting base64 string is stripped per [RFC 7636](https://datatracker.ietf.org/doc/html/rfc7636#section-4.2).

### MD5

```go
func MD5(text string) string
```

`MD5` creates an MD5 hash from the provided plain text.

### SHA256

```go
func SHA256(text string) string
```

`SHA256` creates a SHA-256 hash as defined in FIPS 180-4 from the provided text.

### SHA512

```go
func SHA512(text string) string
```

`SHA512` creates a SHA-512 hash as defined in FIPS 180-4 from the provided text.

### HS256

```go
func HS256(text string, secret string) string
```

`HS256` creates an HMAC hash with the SHA-256 digest algorithm, using the provided secret.

### HS512

```go
func HS512(text string, secret string) string
```

`HS512` creates an HMAC hash with the SHA-512 digest algorithm, using the provided secret.

### Equal

```go
func Equal(hash1 string, hash2 string) bool
```

`Equal` compares two hash strings for equality without leaking timing information.

### HashWithSalt

```go
func HashWithSalt(value string) string
```

`HashWithSalt` hashes the provided value using SHA-256 and a predefined salt value.

### GenerateJWT

```go
func GenerateJWT(claims jwt.Claims, secretKey string) (string, error)
```

`GenerateJWT` generates a JSON Web Token (JWT) with the given claims and secret key, using the HS256 signing method.

### ValidateJWT

```go
func ValidateJWT(tokenString string, secretKey string, claims jwt.Claims) error
```

`ValidateJWT` validates the given JWT token string using the provided secret key and populates the claims if the token is valid. It expects the token to be signed with the HS256 signing method.

### GenerateRandomString

```go
func GenerateRandomString(length int) string
```

`GenerateRandomString` generates a random string of the specified length, using a cryptographically secure random number generator. The string consists of uppercase and lowercase letters, and digits.

## Usage

Import the package in your Go code:

```go
import "github.com/your-org/your-repo/security"
```

Then, you can use the provided functions as needed. For example:

```go
// Generate a SHA-256 hash
hash := security.SHA256("hello world")

// Create an HMAC hash with SHA-512
hmac := security.HS512("secret message", "my-secret-key")

// Generate a JWT token
claims := jwt.MapClaims{
    "sub": "user123",
    "exp": time.Now().Add(time.Hour * 24).Unix(),
}
token, err := security.GenerateJWT(claims, "my-secret-key")
if err != nil {
    // Handle error
}

// Generate a random string
randomString := security.GenerateRandomString(16)
```

Please refer to the documentation and source code for more detailed information on each function and its usage.
