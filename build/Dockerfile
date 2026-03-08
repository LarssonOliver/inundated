ARG VERSION="dev"

FROM --platform=$BUILDPLATFORM node:24-alpine AS frontend-builder

WORKDIR /app

COPY frontend/package*.json .

RUN npm install

RUN sed -i "s/\"version\": \".*\"/\"version\": \"${VERSION}\"/" package.json

COPY frontend/ .

RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.25 AS backend-builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY --from=frontend-builder /app/*.go ./frontend/
COPY --from=frontend-builder /app/dist ./frontend/dist/

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -tags embed \
    -ldflags="-s -w -X cmd/server/main.Version=${VERSION}" \
    -o inundated \
    ./cmd/server/main.go

FROM gcr.io/distroless/static:nonroot

COPY --from=backend-builder /app/inundated /inundated

ENTRYPOINT ["/inundated"]

