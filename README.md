# Go Fiber Real-Time Chat Application

![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)
![Framework](https://img.shields.io/badge/Fiber-v2-00ADD8?style=for-the-badge&logo=go)
![Database](https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker)
![License](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)

A high-performance, feature-rich real-time chat application built with Go (Fiber) on the backend and Vanilla JavaScript on the frontend. This project demonstrates a clean backend architecture, WebSocket-based real-time messaging, and secure JWT-based authentication.

---

### Replace This Screenshot!

_Update the URL below with your app’s actual screenshot._

![App Screenshot](https://via.placeholder.com/800x450.png?text=Your+App+Screenshot)

---

## ✨ Features

- **Secure Authentication:** Registration, login, logout using JWT (Access & Refresh Tokens).
- **Multi-Room Chat:**
  - **Direct Messages (DM):** Auto-created 1-on-1 private chat rooms.
  - **Group Chats:** Create or join group chats with multiple participants.
- **Real-Time Messaging:** Instant message delivery via WebSocket.
- **Lobby & Chat History:** View joined rooms and load previous messages on entry.
- **Persistent Database:** PostgreSQL for users, rooms, and message storage.
- **Dynamic UI:** Intuitive frontend with auto-generated avatars and readable user display names.

## 📦 Tech Stack

| Component         | Technology                                                                                                                          |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| **Backend**       | Go, [Fiber v2](https://gofiber.io/), [GORM](https://gorm.io/), [Google Wire](https://github.com/google/wire), PostgreSQL, WebSocket |
| **Frontend**      | Vanilla JavaScript (ES6+), HTML5, CSS3                                                                                              |
| **Deployment**    | Docker, Docker Compose                                                                                                              |
| **Cloud Storage** | [Cloudinary](https://cloudinary.com/) for image uploads and CDN delivery                                                            |

---

## 📱 Mobile Frontend

> This project is compatible with a Flutter-based mobile frontend available here:

**🔗 [Flutter Chat App Repository](https://github.com/Kirara02/Flutter-ChatApp.git)**

---

## 🧱 Architecture

This project follows a **layered architecture** to ensure clean separation of concerns and maintainability:

```
+----------------+      +----------------+      +------------------+      +-------------+
|     Client     |----->|     Handler    |----->|      Service     |----->| Repository  |
| (JS/WebSocket) |      | (Validation)   |      | (Business Logic) |      | (Database)  |
+----------------+      +----------------+      +------------------+      +-------------+
```

- **Handler:** Receives HTTP/WebSocket requests and performs validation.
- **Service:** Contains application business logic.
- **Repository:** Handles communication with the database.
- **Dependency Injection:** Managed via [Google Wire](https://github.com/google/wire).

---

## ☁️ Cloudinary Integration

This project uses **Cloudinary** for hosting and delivering profile images with CDN support.

### Why Cloudinary?

- Automatic image optimization and resizing
- High availability via CDN
- Public ID–based upload and retrieval
- Built-in image foldering (e.g., `go-chat-app/profiles`)

### Required `.env` Configuration

```env
# Cloudinary Configuration
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
CLOUDINARY_BASE_FOLDER=go-chat-app
```

---

## 🚀 Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) (v1.24+)
- [Docker](https://www.docker.com/get-started) & [Docker Compose](https://docs.docker.com/compose/install/)

---

### Method 1: Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/your-username/go-fiber-chatapp.git
cd go-fiber-chatapp

# Create a .env file
cp .env.example .env
# Fill in your database, JWT, and Cloudinary credentials

# Build and run
docker-compose up --build
```

Access the app at [http://localhost:8080](http://localhost:8080)

---

### Method 2: Manual Setup (Without Docker)

1. Clone the repo and set up `.env` as shown above.
2. Make sure PostgreSQL is running and create the necessary database.
3. Install Go dependencies:

```bash
go mod tidy
```

4. Generate Dependency Injection files:

```bash
go generate ./...
```

5. Run the app:

```bash
go run .
```

---

## 🔌 API Endpoints

| Method | Endpoint             | Description                           |
| ------ | -------------------- | ------------------------------------- |
| POST   | `/api/auth/register` | Register a new user                   |
| POST   | `/api/auth/login`    | Login and get access + refresh tokens |
| POST   | `/api/auth/logout`   | Invalidate refresh token              |
| POST   | `/api/auth/refresh`  | Refresh access token                  |
| GET    | `/api/profile`       | Get current user's profile            |
| PUT    | `/api/profile`       | Update current user's profile         |
| GET    | `/api/users`         | List all users                        |
| GET    | `/api/rooms`         | List rooms joined by user             |
| POST   | `/api/rooms`         | Create a new room (group or DM)       |
| GET    | `/chat/ws/:roomId`   | WebSocket endpoint for a room         |

---

## 🛣️ Roadmap

- [ ] Online/Offline user status
- [ ] Typing indicators
- [ ] Unit and integration tests
- [ ] Group management (invite/kick)
- [ ] File and image attachments in chat

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
