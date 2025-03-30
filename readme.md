# Redis Transaction Demo

This project demonstrates the use of Redis transactions with a React frontend and a Gin backend. Users can interact with a simple UI to perform various Redis operations and see the results.

## Project Structure

- **Frontend:** ReactJS application for the user interface.
- **Backend:** Gin server handling API requests and interacting with Redis.

## Prerequisites

- Node.js and npm
- Go (Golang)
- Redis server

## Setup Instructions

### Backend (Gin Server)

1. **Navigate to the backend directory:**

   ```bash
   cd backend
   ```

2. **Initialize the Go module and install dependencies:**

   ```bash
   go mod init github.com/yourusername/redis-demo/backend
   go get -u github.com/gin-gonic/gin
   go get -u github.com/go-redis/redis/v8
   ```

3. **Run the Gin server:**

   ```bash
   go run main.go
   ```

### Frontend (ReactJS)

1. **Navigate to the frontend directory:**

   ```bash
   cd frontend
   ```

2. **Install dependencies:**

   ```bash
   npm install
   ```

3. **Start the React app:**

   ```bash
   npm start
   ```

### Redis Server

Ensure that your Redis server is running. You can start it with:

```bash
redis-server
```

## Usage

1. **Open the React app in your browser:**

   Navigate to `http://localhost:3000` to access the UI.

2. **Interact with the UI:**

3. **Check the Response:**

   The response from the backend will be displayed below the buttons, showing the result of the transaction.

## Redis Transaction Principles

Redis is single-threaded and all commands in a transaction are executed sequentially. This guarantees that the commands are executed as a single isolated operation.

```bash
From the Redis documentation:

All the commands in a transaction are serialized and executed sequentially. A request sent by another client will never be served in the middle of the execution of a Redis Transaction. This guarantees that the commands are executed as a single isolated operation.
```

Redis transactions are atomic, with no rollback in case of error:

```bash
From the Redis documentation:

Redis does not support rollbacks of transactions since supporting rollbacks would have a significant impact on the simplicity and performance of Redis.
```

## Redis Transaction Cases

This demo explores several cases to show how transactions work in Redis:

### Case 1: Syntax Error in Transaction

If any operation has a syntax error, the transaction will be aborted and none of the operations will be executed.

```mermaid
sequenceDiagram
    participant Client
    participant Redis
    
    Note over Client, Redis: Start Transaction
    Client->>Redis: MULTI
    Redis-->>Client: OK
    
    Client->>Redis: SET key1 "value1"
    Redis-->>Client: QUEUED
    
    Client->>Redis: SET key2 "value2"
    Redis-->>Client: QUEUED
    
    Note over Client, Redis: Syntax Error in Command
    Client->>Redis: INCR key3 "not-a-number"
    Redis-->>Client: Error - ERR wrong number of arguments for 'incr' command
    
    Note over Client, Redis: Transaction State After Error
    Client->>Redis: SET key4 "value4"
    Redis-->>Client: Error - EXECABORT Transaction discarded because of previous errors
    
    Client->>Redis: EXEC
    Redis-->>Client: Error - EXECABORT Transaction discarded because of previous errors
    
    Note over Client, Redis: Result: No keys were modified
```

### Case 2: Logical Error in Transaction

If any operation has an error in the logic, the transaction will still execute and all other operations will be executed.

```mermaid
sequenceDiagram
    participant Client
    participant Redis
    
    Note over Client, Redis: Initial State
    Client->>Redis: SET counter "hello"
    Redis-->>Client: OK
    
    Note over Client, Redis: Start Transaction
    Client->>Redis: MULTI
    Redis-->>Client: OK
    
    Client->>Redis: SET key1 "value1"
    Redis-->>Client: QUEUED
    
    Note over Client, Redis: Logical Error (trying to INCR a string)
    Client->>Redis: INCR counter
    Redis-->>Client: QUEUED
    
    Client->>Redis: SET key2 "value2"
    Redis-->>Client: QUEUED
    
    Note over Client, Redis: Execute Transaction
    Client->>Redis: EXEC
    Redis-->>Client: [OK, (error) ERR value is not an integer or out of range, OK]
    
    Note over Client, Redis: Result: key1 and key2 were set, INCR failed
    Client->>Redis: GET key1
    Redis-->>Client: "value1"
    
    Client->>Redis: GET key2
    Redis-->>Client: "value2"
    
    Client->>Redis: GET counter
    Redis-->>Client: "hello" (unchanged)
```

### Case 3: Watch and Optimistic Locking

Using the WATCH command to implement optimistic locking. If a watched key is modified by another client, the transaction will be aborted.

```mermaid
sequenceDiagram
    participant Client A
    participant Redis
    participant Client B
    
    Note over Client A, Redis: Client A watches key "stock"
    Client A->>Redis: WATCH stock
    Redis-->>Client A: OK
    
    Client A->>Redis: GET stock
    Redis-->>Client A: "5" (current stock value)
    
    Note over Client B, Redis: Concurrent modification!
    Client B->>Redis: SET stock "0"
    Redis-->>Client B: OK
    
    Note over Client A, Redis: Start Transaction (because stock > 0)
    Client A->>Redis: MULTI
    Redis-->>Client A: OK
    
    Client A->>Redis: DECR stock
    Redis-->>Client A: QUEUED
    
    Note over Client A, Redis: Execute Transaction
    Client A->>Redis: EXEC
    Redis-->>Client A: (nil) - Transaction aborted
    
    Note over Client A, Redis: Transaction aborted because watched key was modified
```

## Project Details

I've implemented these transaction cases in the backend code, with endpoints to demonstrate each case:

- `/txpipeline` - Shows a Redis WATCH transaction with concurrent modification (Case 3)
- `/syntax-error` - Demonstrates Case 1 (syntax error aborts the entire transaction)
- `/logic-error` - Demonstrates Case 2 (logical error allows other commands to execute)

You can access these endpoints via the API to see how Redis transactions behave in each scenario.

## License

This project is licensed under the MIT License.