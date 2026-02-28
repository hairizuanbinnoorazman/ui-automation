FROM node:22-slim

RUN npx -y playwright install --with-deps chromium
RUN npm install -g @playwright/mcp

EXPOSE 3000

CMD ["npx", "@playwright/mcp", "--port", "3000", "--host", "0.0.0.0", "--headless", "--browser", "chrome"]
