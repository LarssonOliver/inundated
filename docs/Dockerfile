FROM node:24-alpine AS builder

WORKDIR /app

COPY docs/package*.json .

RUN npm ci

COPY docs/ .

RUN npm run build

FROM nginxinc/nginx-unprivileged:1.27-alpine

COPY --from=builder /app/build /usr/share/nginx/html
COPY docs/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 8080
