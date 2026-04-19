# P2 - Advanced Media Pipeline (Go)

This project is the second phase of the Go Learning Path. It focuses on handling heavy async workloads, file streaming, cloud storage integration, and background job processing using Redis.

## 🎯 Goal
Build a "Media Hub" where users can upload high-resolution profile images. The API will immediately respond to the user while background workers handle heavy image processing asymptotically based on user requirements (validation, stripping EXIF data, resizing, compressing, or preserving original quality).

## 🏗️ Architecture Split
1. **HTTP API Server**: Handles immediate user requests, validates file uploads, processes user processing parameters (e.g. compress, improve, keep-as-is), applies rate limiting, initially saves raw files to AWS S3, and queues background tasks.
2. **Background Worker (Asynq/Redis)**: Listens to the task queue. Picks up image processing tasks, dynamically applies requested transformations using Go image packages or C-bindings, uploads the finished assets back to S3, and updates the database status.

## 🚀 Key Features
- **Rate-Limited Uploads**: Prevent abuse using IP or User-ID based rate limiting (`golang.org/x/time/rate`).
- **Multipart Form Parsing**: Efficiently parse large file uploads without crashing the server's RAM.
- **S3 Integration**: Stream files directly to/from AWS S3 buckets using `aws-sdk-go-v2`.
- **Advanced Image Processing**: 
  - **Validation & Decoding**: Ensure the uploaded file is a valid image format (JPEG, PNG).
  - **Metadata Stripping**: Remove EXIF data (e.g., GPS coordinates, camera info) to protect user offline privacy locally before compressing.
  - **Dynamic Manipulation based on Input**: 
    - *Compress/Resize*: Center-crop to a perfect square, downscale to target resolutions (e.g., 512x512, 128x128), and convert to WebP or high-compression JPEG.
    - *Quality Improve/Keep Original*: Apply minor sharpening filter and retain full resolution.
- **Asynchronous Task Queue**: Robust, persistent background jobs with built-in retries using `hibiken/asynq` and Redis.

## 🛠️ Required Tech Stack
- **Web Framework**: `go-chi/chi/v5`
- **Database**: PostgreSQL (with `sqlc` for type-safe queries) + `pgx`
- **Task Queue**: `hibiken/asynq` (Requires Redis)
- **Object Storage**: AWS SDK `github.com/aws/aws-sdk-go-v2`
- **Rate Limiting**: `golang.org/x/time/rate`
- **Image Processing**: `github.com/h2non/bimg` or standard `image`, `image/jpeg`, `image/draw` packages.

## 📋 Prerequisites
Ensure the following are installed and running locally:
1. **PostgreSQL**: For storing user and media metadata.
2. **Redis**: For the `asynq` task queue and rate-limiting state.
3. **AWS Account**: An S3 bucket and IAM user credentials (or MinIO for local S3 simulation).

---

Read the [Step-by-Step Plan](step_by_step_plan.md) to start building!
