# Threads Clone Backend

Kloning aplikasi Threads dari Meta yang dibangun menggunakan arsitektur microservices dengan Go dan Python.

## 🏗️ Arsitektur

Sistem ini terdiri dari beberapa microservices:

- **API Gateway** (Port 3000) - Entry point utama
- **Auth Service** (Port 3001) - Manajemen user dan autentikasi
- **Threads Service** (Port 3002) - Manajemen threads dan likes
- **Follow Service** (Port 3003) - Manajemen follow/unfollow
- **AI Service** (Python) - Moderasi konten otomatis

## 🛠️ Tech Stack

- **Backend Framework**: Fiber (Go)
- **Database**: PostgreSQL dengan sqlx
- **Message Queue**: RabbitMQ
- **Authentication**: JWT
- **Containerization**: Docker & Docker Compose

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL (jika tidak menggunakan Docker)

### Setup Development Environment

1. **Clone repository**
```bash
git clone <repository-url>
cd threads-clone-backend
```

2. **Setup setiap service**
Jalankan script setup yang telah disediakan:
```bash
# Jalankan semua commands dari artifact "Project Setup Commands"
```

3. **Konfigurasi Environment Variables**
Copy file `.env.example` ke `.env` di setiap service dan sesuaikan nilai-nilainya.

4. **Start dengan Docker Compose**
```bash
docker-compose up --build
```

### Manual Setup (tanpa Docker)

1. **Start PostgreSQL dan RabbitMQ**
```bash
# PostgreSQL
createdb threads_db
psql threads_db < init.sql

# RabbitMQ
sudo systemctl start rabbitmq-server
```

2. **Start setiap service secara terpisah**
```bash
# Terminal 1 - Auth Service
cd auth-service
go run cmd/main/main.go

# Terminal 2 - Threads Service  
cd threads-service
go run cmd/main/main.go

# Terminal 3 - Follow Service
cd follow-service
go run cmd/main/main.go

# Terminal 4 - API Gateway
cd api-gateway
go run cmd/main/main.go
```

## 📊 Database Schema

```sql
-- Users
users: id(UUID), username, display_name, email, password_hash, bio, profile_image_url, created_at

-- Threads  
threads: id(UUID), user_id, content, parent_thread_id, status, created_at

-- Likes
likes: user_id, thread_id, created_at

-- Follows
follows: follower_id, following_id, created_at
```

## 🔧 API Endpoints

### Auth Service (Port 3001)
- `POST /auth/register` - Register user baru
- `POST /auth/login` - Login user
- `GET /users/profile` - Get user profile
- `PUT /users/profile` - Update user profile

### Threads Service (Port 3002)
- `POST /threads` - Buat thread baru
- `GET /threads` - Get threads (feed)
- `GET /threads/:id` - Get thread by ID
- `POST /threads/:id/like` - Like/unlike thread
- `POST /threads/:id/reply` - Reply to thread

### Follow Service (Port 3003)
- `POST /follow/:userId` - Follow user
- `DELETE /follow/:userId` - Unfollow user
- `GET /following` - Get users yang di-follow
- `GET /followers` - Get followers

### API Gateway (Port 3000)
- Semua endpoint di atas dapat diakses melalui gateway dengan prefix service
- `GET /health` - Health check semua services

## 🧪 Testing

```bash
# Run tests untuk semua services
make test

# Run tests untuk service tertentu
cd auth-service && go test ./...
```

## 🐳 Docker Commands

```bash
# Build semua services
docker-compose build

# Start services
docker-compose up

# Stop services  
docker-compose down

# View logs
docker-compose logs -f [service-name]

# Restart specific service
docker-compose restart auth-service
```

## 📁 Project Structure

```
threads-clone-backend/
├── auth-service/        # User management & authentication
├── threads-service/     # Threads & likes management  
├── follow-service/      # Follow/unfollow management
├── api-gateway/         # Request routing & rate limiting
├── docker-compose.yml   # Docker orchestration
└── init.sql            # Database initialization
```

## 🔐 Environment Variables

Setiap service memerlukan file `.env` dengan konfigurasi yang sesuai. Lihat file `.env.example` di setiap service untuk template.

## 🤝 Contributing

1. Fork repository
2. Create feature branch
3. Commit changes
4. Push to branch  
5. Create Pull Request

## 📝 License

MIT License - lihat file LICENSE untuk detail lengkap.