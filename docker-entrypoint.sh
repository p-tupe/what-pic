#!/bin/sh
# set -e is intentionally omitted around the pull pipe —
# grep returns exit 1 on no-match which would kill the script.

OLLAMA_HOST="${OLLAMA_HOST:-http://localhost:11434}"
MODEL="${OLLAMA_MODEL:-llava}"

echo "Waiting for Ollama at ${OLLAMA_HOST}..."
until curl -sf "${OLLAMA_HOST}/api/tags" >/dev/null 2>&1; do
  sleep 2
done

# Pull model if not present. In compose this is a no-op (model-pull service
# already ran). Kept here as a fallback for standalone (non-compose) use.
if ! curl -sf "${OLLAMA_HOST}/api/show" \
     -H "Content-Type: application/json" \
     -d "{\"name\":\"${MODEL}\"}" >/dev/null 2>&1; then
  echo "Pulling ${MODEL} (first run, may take several minutes)..."
  curl -s "${OLLAMA_HOST}/api/pull" \
       -H "Content-Type: application/json" \
       -d "{\"name\":\"${MODEL}\"}" \
    | grep -o '"status":"[^"]*"' \
    | cut -d'"' -f4 \
    | uniq \
    || true
  echo "Model ready."
fi

exec /app/what-pic
