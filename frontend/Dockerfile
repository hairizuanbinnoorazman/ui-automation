# Stage 1: Build Elm application
FROM node:18-alpine AS builder

# Install Elm binary directly from GitHub
RUN wget -O elm.gz https://github.com/elm/compiler/releases/download/0.19.1/binary-for-linux-64-bit.gz \
    && gunzip elm.gz \
    && chmod +x elm \
    && mv elm /usr/local/bin/

WORKDIR /app
COPY elm.json ./
COPY src ./src
COPY index.html ./
RUN elm make src/App.elm --output=elm.js --optimize

# Stage 2: Nginx server
FROM nginx:alpine
COPY --from=builder /app/elm.js /usr/share/nginx/html/
COPY --from=builder /app/index.html /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
