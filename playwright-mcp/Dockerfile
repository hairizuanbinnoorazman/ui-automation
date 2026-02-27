FROM node:22-slim

RUN npx -y playwright install --with-deps chromium
RUN npm install -g @anthropic-ai/playwright-mcp

EXPOSE 3000

CMD ["npx", "@anthropic-ai/playwright-mcp", "--port", "3000", "--headless", "--browser", "chromium"]
