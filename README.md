# JWT
JWT provides a compact, self-contained way to securely transmit information as a JSON object.

#### JWT Structure
A JWT consists of three parts encoded in Base64URL format and separated by dots:
```
Header.Payload.Signature
```

For example:
```
eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiZXhwIjoxNjgwMDAwMDAwfQ.8Gj_9bJjAqQ-5j3iCKMzVnlg-d1Kk-fXnOKC1Vt2fGc
```

1. The header identifies the algorithm used for signing:
```
// In Go, the header is typically handled by the JWT library
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
```

2. The payload contains claims about the user like ID, roles, and expiration time
```
// Creating claims in Go
claims := jwt.MapClaims{
    "sub": user.ID.String(),
    "username": user.Username,
    "exp": time.Now().Add(15 * time.Minute).Unix(),
}
```

3. The signature verifies the token hasn't been tampered with
```
// Signing the token with our secret
tokenString, err := token.SignedString([]byte(jwtSecret))
```

# Project Setup 
#### Structure
```
go-jwt-mysql-lab/
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── auth/
│   │   ├── jwt.go
│   │   └── middleware.go
│   │
│   ├── database/
│   │   └── mysql.go
│   │
│   ├── user/
│   │   ├── model.go
│   │   ├── repository.go
│   │   └── service.go
│   │
│   └── http/
│       ├── handlers.go
│       └── routes.go
│
├── migrations/
│   └── 001_create_users.sql
│
├── .env
├── .gitignore
├── go.mod
└── go.sum
```

#### GO Packages
```
go get github.com/go-sql-driver/mysql 
go get github.com/golang-jwt/jwt/v5 
go get golang.org/x/crypto/bcrypt 
go get github.com/joho/godotenv
```

#### Get MySQL Server (8.4 lts - Linux)
```
https://dev.mysql.com/doc/refman/8.4/en/linux-installation-apt-repo.html#repo-qg-apt-upgrading
```
1. Download MySQL APT repository
```
wget https://dev.mysql.com/get/mysql-apt-config_0.8.36-1_all.deb
```
2. Install the downloaded release package
```
sudo dpkg -i /PATH/version-specific-package-name.deb
```
3. Update το apt
```
sudo apt-get update
```
4. Install MySQL
sudo apt-get install mysql-server

#### Set up Database
1. Create DB
```
CREATE DATABASE gojwt
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;
```

2. Create Application User
```
CREATE USER 'gojwt'@'localhost'
IDENTIFIED BY 'Password123!';
```

3. Grant Privileges
```
GRANT ALL PRIVILEGES ON gojwt.* TO 'gojwt'@'localhost';
FLUSH PRIVILEGES;
```

### Database Migration
#### Create Table Users
The users table is defined in:
migrations/001_create_users.sql

```
mysql -u gojwt -p gojwt < migrations/001_create_users.sql
```

#### Database connection 
Database connection is implemented in:
internal/database/mysql.go

The application uses Go's database/sql package with the MySQL driver.
The connection configuration is loaded from environment variables.

#### .env
```
APP_PORT=8080

MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_DATABASE=gojwt
MYSQL_USER=gojwt
MYSQL_PASSWORD=CHANGE_ME

JWT_SECRET=CHANGE_THIS_TO_A_LONG_RANDOM_SECRET
JWT_ISSUER=go-jwt-lab
JWT_EXPIRATION_MINUTES=15
```

Generate a random JWT secret with:
```
openssl rand -hex 32
```

#### HTTP Server
The HTTP server is implemented in:
cmd/server/main.go

Start the server with:
```
go run ./cmd/server
```

Endpoint:
GET /health

Test:
```
curl -i http://localhost:8080/health
```

Expected response:
```
{
  "status": "ok"
}
```

#### Authentication Flow
The application implements the following authentication flow:
```
POST /register
      │
      ▼
    Password
      │
      ▼
    bcrypt
      │
      ▼
    MySQL
```

Login:
```
POST /login
      │
      ▼
    Find user
      │
      ▼
    bcrypt verification
      │
      ▼
    Generate JWT
```

Protected requests:
```
Authorization: Bearer <JWT>
      │
      ▼
    JWT Middleware
      │
      ▼
    Validate JWT
      │
      ▼
    Extract Claims
      │
      ▼
    Protected Handler
```

#### Register
Endpoint:
POST /register

Example:
```
curl -i \
  -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'
```
Expected response:
```
{
  "id": 1,
  "email": "test@example.com"
}
```
Passwords are never stored as plaintext.
They are hashed using bcrypt before being stored in MySQL.

#### Login
Endpoint:
POST /login

Example:
```
curl -i \
  -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'
```
Expected response:
```
{
  "token": "<JWT>"
}
```
The JWT is signed using HS256.
The token contains claims such as:
```
user_id
email
iss
sub
iat
exp
```

#### Bearer Authentication
Protected endpoints use the standard HTTP Authorization header:

Authorization: Bearer <JWT>

Example:
```
curl -i \
  -H "Authorization: Bearer <JWT>" \
  http://localhost:8080/me
```

#### Protected Endpoints
GET /me

Requires a valid JWT.
```
curl -i \
  -H "Authorization: Bearer <JWT>" \
  http://localhost:8080/me
```

Example response:
```
{
  "email": "test@example.com",
  "id": 1
}
```
Without a token:
```
curl -i http://localhost:8080/me
```
Expected:
```
401 Unauthorized
```

#### GET /protected
Requires a valid JWT.
```
curl -i \
  -H "Authorization: Bearer <JWT>" \
  http://localhost:8080/protected
```
Example response:
```
{
  "message": "you have access to this protected resource",
  "user_id": 1
}
```
Without a token:
```
curl -i http://localhost:8080/protected
```
Expected:
```
401 Unauthorized
```
