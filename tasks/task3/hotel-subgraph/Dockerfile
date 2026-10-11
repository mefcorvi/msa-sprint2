FROM node:24-alpine
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY tsconfig.json ./
COPY src ./src
RUN npm run typecheck && npm prune --omit=dev
# Set after the build: npm ci needs the dev dependencies. In production Apollo omits stack traces from errors.
ENV NODE_ENV=production
USER node
EXPOSE 4002
CMD ["node", "src/index.ts"]
