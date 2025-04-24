# Rahanna

A peer-to-peer (P2P) chess game with a Terminal User Interface (TUI).


https://github.com/user-attachments/assets/cf3409b2-bdfa-49f8-997a-252c584dee1a


> _Disclaimer:_
> This project is a university exercise for a Distributed Systems class. While it seems to function correctly, Rahanna is not intended for production use. The UI opens network ports that should not be exposed in secure environments.

---

Rahanna enables two players to play chess directly via a TCP connection — no centralized server manages the game state. Moves are sent in real time over the network, and only the final outcome is stored.

- P2P gameplay over TCP.
- TUI-based interface.
- Lightweight REST API for user authentication and matchmaking.
- No need to know your opponent's IP address - just the unique game name.

Even though Rahanna is a P2P application, it relies on a central Rahanna API for:

- User authentication and registration.
- Match discovery.

No gameplay data is stored centrally — only metadata (e.g., game outcomes).

## Build

Make sure you have Go 1.24+ and Git installed.

```
$ git clone github.com/boozec/rahanna.git
$ # make all or
$ go build -o rahanna-ui cmd/ui/main.go
```

If you want to also makes up an API server, run

```
$ go build -o rahanna-api cmd/api/main.go
```

## Run

Now, you can just run the one (or two) executables you just builded after an environment setup:

```
export API_BASE="http://localhost:8080"
```

Or, if you also want to make up the API:

```
export POSTGRES_USER=postgres
export POSTGRES_DB=rahanna
export POSTGRES_PASSWORD=password
export DATABASE_URL="host=localhost user=postgres password=password dbname=rahanna port=5432"
export JWT_TOKEN="..."
export RAHANNA_API_ADDRESS=":8080"
export DEBUG=1
```

If you are more lazy, you just can make everything up thanks to Docker.

```
# Start the API and database
docker compose -f docker/api/docker-compose.yml up

# In another terminal, run the UI
docker run -e API_BASE=http://0.0.0.0:8080 -it --network host rahanna-ui:latest
```
