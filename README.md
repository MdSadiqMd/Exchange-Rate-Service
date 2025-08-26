# Exchange-Rate-Service

## Features

- Fetch latest exchange rates from exchangerate.host  
- Convert amounts between multiple currencies 
- In-memory caching (TTL configurable) 
- Health check endpoint  
- Dockerized for easy deployment  

## Prerequisites

- Go 1.21 or higher  
- Docker & Docker Compose (optional)  

## Getting Started

1. Clone the repository  
   ```bash
   git clone https://github.com/MdSadiqMd/Exchange-Rate-Service.git
   cd Exchange-Rate-Service
   ```

2. Update configuration in `config.yaml`
   ```bash
   cp config.example.yaml config.yaml
   ```
   Replace the empty string with API_KEY

3. Run the service locally:  
   ```bash
   go mod tidy
   go run cmd/server/main.go
   ```
   The service listens on port 8080 by default

5. (Optional) Run with Docker Compose:  
   ```bash
   docker-compose up --build
   ```
   - Service: http://localhost:8080 

## API Endpoints

### Health Check  
Basic service health endpoint.
```
curl -X GET "http://localhost:8080/health"
```
Response:
```json
"OK"
```

### Convert Currency  
Convert an amount from one currency to another (optional date within last 90 days).
```
curl -X GET "http://localhost:8080/api/v1/convert?from=USD&to=INR&amount=100&date=2025-08-21"
```
Response:
```json
{
  "success":true,
  "result":8729.365
}
```

## Testing

Run unit and integration tests:
```bash
go test ./... -v
```

## Assumptions
- Only 5 currencies supported: USD, EUR, GBP, JPY, INR 
- Historical rates limited to last 90 days  
- In-memory cache sufficient for single-instance deployment  
- No external DB or Redis; simple in-memory cache  
- Date format strictly `YYYY-MM-DD`  
- Go-kit for endpoint wiring; Chi for HTTP routing  
