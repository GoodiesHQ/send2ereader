# send2ereader

A lightweight Go backend for transferring files from a computer to Kobo or Kindle e-readers using a short temporary code. The app uses Go's embedded HTML templates and HTMX for a single-language frontend/backend experience.

## Features

- Detects Kobo and Kindle user agents
- Generates a temporary code for e-readers (set via env var `CODE_SIZE`, default: 4)
- Upload form for desktop browsers and other devices
- HTMX-powered polling for e-reader status updates
- In-memory ephemeral transfer sessions, no disk usage
- No external database required

## Settings (Environment Variables)

| Name | Default | Purpose |
| ---- | ------- | ------- |
| PORT | 8080 | The port which is utilized by the HTTP server |
| CODE_SIZE | 4 | The number of digits in each session code |
| EXPIRATION_MINUTES | 15 | The number of minutes each file will be temporarily stored |


## Run Locally

```bash
git clone https://github.com/goodieshq/send2ereader.git
cd send2ereader
go mod tidy
go run .
```

Then open `http://localhost:8080`.

## Run via Docker

```bash
docker run -p 8080:8080 goodieshq/send2ereader:latest
```

## Run via Docker Compose

Minimal:

```yaml
services:
  send2ereader:
    image: goodieshq/send2ereader:latest
    container_name: send2ereader
    restart: unless-stopped
    pull_policy: always
    network_mode: "host"
    ports:
      - "8080:8080"
```

Traefik Example (using an external network called "frontend"):

```yaml
services:
  send2ereader:
    image: goodieshq/send2ereader:latest
    container_name: send2ereader
    restart: unless-stopped
    pull_policy: always
    networks:
      - frontend
    environment:
        PORT: 8080
        CODE_SIZE: 4
        EXPIRATION_MINUTES: 15
    labels:
      traefik.enable: "true"
      traefik.docker.network: frontend
      traefik.http.routers.send2ereader.rule: Host(`s2e.example.com`)
      traefik.http.routers.send2ereader.entrypoints: websecure
      traefik.http.routers.send2ereader.tls: "true"
      traefik.http.routers.send2ereader.tls.certresolver: my_resolver
      traefik.http.services.send2ereader.loadbalancer.server.port: 8080
      traefik.http.services.send2ereader.loadbalancer.server.scheme: http
      traefik.http.routers.send2ereader.service: send2ereader
networks:
  frontend:
    external: true

```


## Notes
- Kepubify only apples when:
  - The original file is a `.epub` (and not already a `.kepub.epub`)
  - The download request is coming from a device with a Kobo user agent.