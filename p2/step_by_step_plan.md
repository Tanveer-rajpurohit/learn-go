# P2 Implementation Plan & Architecture

## 🛠️ Required Packages & Tech Stack
- **Routing & HTTP**: `go-chi/chi/v5`
- **Database**: PostgreSQL with `jackc/pgx/v5` and `sqlc`
- **Background Task Queue**: Pure Redis using `redis/go-redis/v9` (Implementing a custom Producer/Consumer queue)
- **Cloud Storage**: AWS S3 via S3 Go SDK v2 (`aws/aws-sdk-go-v2`)
- **Image Processing**: `github.com/h2non/bimg` (wraps C-level libvips for extreme speed) or standard Go `image` package
- **Rate Limiting**: `golang.org/x/time/rate`
- **Environment Config**: `joho/godotenv`

## 🔌 Endpoints Required
1. **`POST /upload/image`** (Synchronous Upload)
   - Parses multipart form data (the image file).
   - Accepts processing options (`compress`, `keep`, `improve`).
   - Validates image magic bytes (ensure it's a real image, not a script).
   - Uploads original file to S3.
   - Saves record to DB and publishes a JSON task payload to a Redis List using `LPUSH`.
   - Returns HTTP 202 Accepted with an `asset_id`.
2. **`GET /media/{id}`** (Client Polling Endpoint)
   - Frontend hits this endpoint every 3-5 seconds.
   - Checks DB for status (`pending`, `processing`, `completed`, `failed`).
   - Returns `{"status": "processing"}` if worker is still running.
   - Returns S3 URLs of the variants if `completed`.

## 🗄️ Database Schema Required
- **Table: `users`** (Assuming basic auth exists from P1)
- **Table: `media_assets`**
  - `id` (UUID, Primary Key)
  - `user_id` (UUID, Foreign Key)
  - `status` (VARCHAR/ENUM: 'pending', 'processing', 'completed', 'failed')
  - `raw_url` (VARCHAR - S3 link to original file)
  - `processing_options` (JSONB - user instructions like "compress", "preserve")
  - `variants` (JSONB - URLs generated after background processing)

## 🚀 Steps to Implement (Chronological)

### Step 1: Project Setup & Docker
- [ ] Initialize `go mod init github.com/yourusername/p2` inside the `p2` directory.
- [ ] Configure `docker-compose.yml` to spin up local PostgreSQL and Redis containers.
- [ ] Install the listed `go get` dependencies.

### Step 2: Database Initialization
- [ ] Write the `schema.sql` for `media_assets`.
- [ ] Generate Go CRUD models using `sqlc generate`.

### Step 3: AWS S3 Service Layer
- [ ] Create `internal/aws/s3.go`.
- [ ] Implement `UploadFile(ctx, file io.Reader, filename string) (string, error)`. 
  - *Note*: Pass `io.Reader` directly to stream the file, avoiding loading heavy images entirely into RAM.

### Step 4: The Synchronous API (Frontend Facing)
- [ ] Set up the Chi router and apply the Token Bucket Rate Limiting middleware.
- [ ] Implement the `POST /upload/image` handler: Save to DB (`status='pending'`). Pack the `AssetID` and `ProcessingOptions` into a JSON struct.
- [ ] Use a `go-redis/v9` client to `LPUSH` (Left Push) the JSON task payload onto an `image_processing_queue` list in Redis.
- [ ] Implement the `GET /media/{id}` handler for the Polling flow.

### Step 5: Background Worker Setup (Native go-redis)
- [ ] Create a new entrypoint for the worker process (e.g., `cmd/worker/main.go`).
- [ ] Initialize the `redis.Client` connecting to your Redis instance.
- [ ] Create an infinite `for` loop that uses `BRPOP` (Blocking Right Pop) acting as the queue consumer on `image_processing_queue`.
  - *Learning Point*: `BRPOP` completely blocks the worker until an item enters the list, meaning zero CPU utilization when idle rather than spamming Redis to check for work.

### Step 6: The Image Processor Worker Logic
- [ ] Unmarshal the JSON payload popped from Redis to get the `AssetID` and `ProcessingOptions`.
- [ ] Set DB status to `processing`.
- [ ] Download raw image from S3.
- [ ] **Strip Metadata**: Parse and strip EXIF data heavily to protect user privacy (GPS coords, device tracking).
- [ ] Execute requested logic:
  - If `compress`: Resize to 512x512, convert to WebP or HTTP-optimized JPEG.
  - If `keep`: Copy over untouched (but without EXIF).
- [ ] Upload final result assets back to S3.
- [ ] Update DB status to `completed` and populate `variants`.

### Step 7: End-to-End Testing (The Polling Simulation)
- [ ] Run the API server and the Worker server simultaneously.
- [ ] Use Postman or cURL to `POST /upload/image` and capture the `asset_id`.
- [ ] Repeatedly query `GET /media/{id}` (simulating frontend polling) every 3 seconds.
- [ ] Verify you see the transition from `pending` -> `processing` -> `completed` and the S3 links are generated!
- [ ] Add a retry mechanism: If processing fails, push the task back to the queue (or a DLQ) instead of letting it drop forever.
