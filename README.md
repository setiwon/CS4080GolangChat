# Go Chat App

This is a simple chat application built with Go programming language using WebSocket for real-time communication.

## Group Members

- Wesley Dam
- Joseph Setiawan
- Amar Gandhi
- Samuel Araya
- Phu Truong

## Features

- Real-time messaging: Users can send and receive messages instantly.
- File sharing: Users can upload and share files with other participants in the chat room.
- Persistence: Chat messages are stored in an SQLite database, allowing users to see previous messages upon joining.

## Installation

1. Clone the repository:

    ```bash
    https://github.com/setiwon/CS4080GolangChat.git
    ```

2. Navigate to the project directory:

    ```bash
    cd CS4080GolangChat
    ```

3. Install dependencies:

    ```bash
    go mod tidy
    ```
    
## Usage

1. Run the server:

    ```bash
    go run .
    ```

2. Open your web browser and navigate to `http://localhost:8080` to access the chat application.

3. Enter your name and start chatting!

## Dependencies

- [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3): SQLite driver for Go.
- [github.com/gorilla/websocket](https://github.com/gorilla/websocket): WebSocket library for Go.
- [Bootstrap](https://getbootstrap.com/): Front-end CSS framework for styling.
