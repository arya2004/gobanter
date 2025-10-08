# GoBanter 💬

GoBanter is a lightweight, real-time chat application built in Go. It uses Gorilla WebSocket for seamless real-time communication, Pat URL pattern muxer for routing, and jQuery for simple front-end interactions. Perfect for learning or deploying a scalable chat system with minimal overhead.

---

## Features
- **Real-time Messaging**: Powered by WebSockets for instant communication.
- **Simple & Lightweight**: Easy to set up and run with minimal dependencies.
- **Cross-Device Support**: Accessible from multiple devices simultaneously.
- **Customizable**: A great foundation to build your chat application.

---

## Getting Started

### Prerequisites
- [Go](https://golang.org/dl/) (1.16 or higher)
- Basic understanding of WebSockets and Go development (helpful but not required).

---

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/arya2004/gobanter.git
   cd gobanter
   ```

2. **Install dependencies**:
   ```bash
   go  mod tidy
   ```

3. **Run the application**:
   ```bash
   sh run.sh
   ```

4. **Access the application**:
   Open your browser and navigate to `http://localhost:8080`.

---

## Usage

1. Open the application in multiple tabs or devices.
2. Start typing messages in the chat interface.
3. Messages will appear in real-time across all connected users.

---

## WebSocket Message Format

GoBanter uses simple JSON messages over the WebSocket. Every message is an object with an "Action" field (string) and a "Payload" field (object). The server and clients rely on the Action value to determine how to handle the Payload.

Supported action types (common examples)
- "username" — client sets or updates their username.
- "broadcast" — client sends a chat message to be broadcast to all connected users; server forwards broadcast events to clients.
- "left" — server notifies clients that a user disconnected.
- "connected" — server notifies clients that a user connected (or server acknowledges a new connection).
- Additional actions may be defined (e.g., "userlist", "history") depending on server extensions.

Common payload fields
- Message (string) — the chat text.
- Username (string) — display name of the sender.
- Conn (string) — connection identifier assigned by the server (useful to disambiguate users).
- Time (string, RFC3339) — optional timestamp included by the server.
- Any other fields are considered extension-specific.

Sample JSON messages

Client -> Server (set username)
```json
{
  "Action": "username",
  "Payload": {
    "Username": "alice"
  }
}
```

Client -> Server (send a chat message)
```json
{
  "Action": "broadcast",
  "Payload": {
    "Message": "Hello, everyone!",
    "Username": "alice"
  }
}
```

Server -> Clients (broadcast received message)
```json
{
  "Action": "broadcast",
  "Payload": {
    "Message": "Hello, everyone!",
    "Username": "alice",
    "Conn": "conn_abc123",
    "Time": "2025-10-06T12:00:00Z"
  }
}
```

Server -> Clients (user left)
```json
{
  "Action": "left",
  "Payload": {
    "Username": "alice",
    "Conn": "conn_abc123"
  }
}
```

Notes
- Implementations should gracefully ignore unknown Action values to remain forward-compatible.
- The server may add extra metadata fields in Payload; clients should only process known fields.

---

## Contributing

We welcome contributions! Whether it's a bug fix, feature addition, or documentation improvement, feel free to open an issue or submit a pull request.

### Steps to Contribute:
1. Fork the repository.
2. Create a feature branch (`git checkout -b feature-name`).
3. Commit your changes (`git commit -m "Add new feature"`).
4. Push to your fork (`git push origin feature-name`).
5. Open a pull request.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---
