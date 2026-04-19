# Go Learning Path & Project Plan

Based on your strong background in Node.js and DevOps, this curriculum is designed to map your existing advanced knowledge (web servers, background jobs, real-time communication, microservices) to idiomatic Go. 

## P1: Foundation & REST APIs (Completed)
- **Goal**: Understand Go's project structure, routing, middle-wares, database communication, and basic security.
- **Tech Stack**: `go-chi`, `pgx`, `sqlc`, `golang-jwt`, `bcrypt`.
- **Key Go Learnings**: Context management, standard library `net/http` fundamentals, struct tags, error handling, DB connection pooling.

---

## P2: Advanced Media Pipeline (File Handling, S3, Background Workers)
**Idea**: "Media Hub" where users can upload high-res profile images or photography assets. The pipeline dynamically processes images based on user preferences (keep original quality, compress to save space, or enhance/resize).

**Normal APIs (Synchronous):**
1. `/upload/image` (POST): Validates image file type, accepts processing options (e.g., `{"mode": "compress", "target_format": "webp"}`), uploads original to S3 directly or via server. Rate Limited.
2. `/download/media/{id}` (GET): Serve processed image files or originals.

**Background Workers (Asynchronous):**
1. **Image Processor Worker**: Depending on the user input from the payload, this worker applies:
   - **Validation & Decoding**: Ensures it's a valid JPEG/PNG.
   - **Metadata Stripping**: Removes EXIF data for privacy.
   - **Resizing & Cropping**: Center-crops or scales based on requirements.
   - **Format Conversion & Compression**: Outputs WebP or compressed JPEG, or keeps original quality as requested. Updates the database profile (`{ variants: [...] }`).

- **Tech Stack**: 
  - `aws/aws-sdk-go-v2` (AWS S3)
  - `hibiken/asynq` (Redis-backed task queue)
  - `gopkg.in/h2non/bimg.v1` (Fast image compression) or standard `image` package
  - `golang.org/x/time/rate` (Rate Limiting)
- **Key Go Learnings**: Managing background job queues, streaming large files without loading the entire payload into RAM (`io.Reader`/`io.Writer`), advanced image manipulation in Go.

---

## P3: Real-Time Multiplayer Drawing Game (WebSockets & Concurrency)
**Idea**: "Skribbl.io Clone" - A real-time multiplayer game where one player gets a word and draws it on a canvas, while other players race to type the correct guess in chat. 

**Normal APIs (Synchronous):**
1. `/games/create` (POST): Initializes a game room, returns a unique ID.
2. `/games/join` (GET/WS): Handles the HTTP Upgrade to a persistent WebSocket connection.

**WebSocket Server (Real-Time):**
1. **Game Loop**: A central goroutine (a "room manager") tracks turns, handles timers (using Go's `time.Ticker`), randomly selects words, and scores players.
2. **Event Broadcaster**: Receives canvas X/Y coordinates from the "drawer" and instantly broadcasts them to all "guessers" in the room.
3. **Chat/Guess Evaluator**: Listens to chat messages; if a guess strictly matches the current target word, it awards points instead of broadcasting the plain text to everyone.

- **Tech Stack**: 
  - `gorilla/websocket` or `nhooyr.io/websocket`
  - `go-redis/redis` (Using Redis Pub/Sub if you want rooms to scale horizontally across multiple instances)
- **Key Go Learnings**: Handling high-throughput concurrent WebSocket connections, Channels vs. Mutexes (`sync.RWMutex`), Select statements, Memory leak prevention (safely reaping goroutines and channels when a user disconnects).

---

## P4: Microservices, gRPC & Distributed Event Streaming
**Idea**: Evolve P2 & P3 into a Distributed System with Video Processing
Introduce heavy media processing (video) through an asynchronous microservice. Refactor P2 and P3 to simulate a high-traffic environment appropriate for your DevOps background.

**Core Architecture Breakdown:**
1. Extract the **Authentication & Profile Service** into an independent gRPC server.
2. Build a **Video Transcoding Service**: Listens for Kafka events for newly uploaded videos, downloads them from S3, and executes `ffmpeg` via Go's `os/exec` to transcode them into multiple qualities (`360p`, `720p`, `1080p`).
3. Introduce a Message Broker (RabbitMQ or Kafka). When a user uploads a video via the API Gateway, publish a `VideoUploaded` event. The Transcoding service consumes this, processes the file, and publishes a `VideoProcessed` event.
4. A dedicated **WebSocket Notification Service** consumes `VideoProcessed` events and instantly alerts the connected user that their video is ready.

- **Tech Stack**: 
  - `grpc/grpc-go` & Protobuf (`protoc`)
  - `rabbitmq/amqp091-go` or `confluentinc/confluent-kafka-go`
  - Standard `os/exec` (Executing CLI FFmpeg)
  - `opentelemetry-go` (for Jaeger/Zipkin Distributed Tracing)
- **Key Go Learnings**: Inter-service communication securely using Protocol Buffers, asynchronous event-driven architectures (EDA), process orchestration (`os/exec` for video processing in a worker), managing distributed latency context (`context.WithTimeout`).

---

## P5: Custom API Gateway & Infrastructure CLI
**Idea**: Reverse Proxy & DevOps Automation Tools
Use Go to build the operational layers that actually sit in front of, and manage, the P4 microservices.

**Core Features:**
1. **API Gateway / Reverse Proxy**: A Go service that listens on an entrypoint (80/443) and dynamically routes requests to the corresponding Microservices (Auth, Game, Video). Use Go's builtin `net/http/httputil.ReverseProxy`.
2. **Custom Gateway Middleware**: Do global Rate Limiting, authentication header attachment, and inject tracing IDs seamlessly on the proxy layer.
3. **CLI Admin Tool**: Build a Command-Line tool so you, as the platform admin, can execute tools programmatically: fetch metrics, purge Redis queues, or forcefully kill an active game room.

- **Tech Stack**: 
  - Standard `net/http/httputil`
  - `spf13/cobra` (Structure for CLI commands like `admin-go get users` or `admin-go purge cache`)
  - `fsnotify/fsnotify` (Live-reloading proxy configurations without downtime)
- **Key Go Learnings**: Low-level networking, modifying HTTP request/response payloads in transit, building production-grade CI/CD and operator tooling, graceful shutdowns (`http.Server#Shutdown`).
