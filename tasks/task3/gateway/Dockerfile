FROM node:24-alpine
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY tsconfig.json ./
COPY src ./src
RUN npm run typecheck && npm prune --omit=dev
ENV NODE_ENV=production
USER node
EXPOSE 4000
CMD ["node", "src/index.ts"]
