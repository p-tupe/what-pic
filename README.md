<div align="center">
  <img src="static/logo.svg" alt="what-pic" width="180"/>

  <h1>what-pic</h1>
  <p>A picture is worth a thousand words; what-pic describes them for you.</p>
</div>

---

## Quickstart

```bash
git clone https://github.com/pritesh/what-pic
cd what-pic
docker compose up
```

Open [http://localhost:8080](http://localhost:8080).

Ollama and the `llava` vision model are set up automatically. The first run downloads the model (~4 GB) — subsequent starts are instant.

---

## How it works

Upload one or more images, write a prompt, pick an output format. EXIF metadata (GPS, camera, datetime) is extracted and injected into the AI context automatically. Results can be downloaded as plain text, JSON, YAML, or CSV.

Everything runs locally. No images leave your machine.

## Configuration

Copy `.env.example` to `.env` and edit before running:

```bash
cp .env.example .env
docker compose up -d
```

| Env var | Default | Description |
|---|---|---|
| `OLLAMA_MODEL` | `llava` | Any Ollama vision model (`llava-llama3`, `llama3.2-vision`, `moondream`…) |
| `PORT` | `8080` | Listen port |
| `MAX_UPLOAD_MB` | `50` | Upload size limit |
| `RATE_LIMIT_IMAGES` | — | Set to any value to enable rate limiting (5 images/hour/IP) |
| `TRUST_PROXY` | — | Set to `1` when behind nginx/Caddy/Traefik to trust `X-Forwarded-For` |

## GPU acceleration

Uncomment the `deploy` block in `docker-compose.yml` and install [nvidia-container-toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html):

```yaml
# docker-compose.yml → ollama service
deploy:
  resources:
    reservations:
      devices:
        - driver: nvidia
          count: all
          capabilities: [gpu]
```

## Running without Docker

Requires [Ollama](https://ollama.com) running locally with a vision model pulled:

```bash
ollama pull llava
go build -o what-pic .
./what-pic
```

## API

`POST /analyze` — multipart/form-data

| Field | Description |
|---|---|
| `files[]` | One or more image files |
| `prompt` | Instruction for the model |
| `format` | `text` \| `json` \| `yaml` \| `csv` |

Response body is the formatted result with `Content-Disposition: attachment` for download.

## Useful commands

```bash
# Start in background
docker compose up -d

# View logs
docker compose logs -f

# Change model (edit .env first)
docker compose up -d --force-recreate what-pic

# Stop
docker compose down

# Stop and delete model cache
docker compose down -v
```
