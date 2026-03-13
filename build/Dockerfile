ARG VERSION="dev" # Set a default version, can be overridden at build time with --build-arg VERSION=your_version

FROM --platform=$BUILDPLATFORM node:24-alpine AS frontend-builder

ARG VERSION

WORKDIR /app

COPY frontend/package*.json .

RUN npm install

COPY frontend/ .

RUN sed -i "s/\"version\": \".*\"/\"version\": \"${VERSION}\"/" package.json

RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS backend-builder

ARG VERSION

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY --from=frontend-builder /app/*.go ./frontend/
COPY --from=frontend-builder /app/dist ./frontend/dist/

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -tags embed \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o inundated \
    ./cmd/server/main.go

FROM gcr.io/distroless/static:nonroot

COPY --from=backend-builder /app/inundated /inundated

ENTRYPOINT ["/inundated"]

