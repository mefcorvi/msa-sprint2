FROM node:24-alpine
WORKDIR /app
COPY task3/booking-subgraph/package.json task3/booking-subgraph/package-lock.json ./
RUN npm ci
COPY task3/booking-subgraph/tsconfig.json ./
COPY task3/booking-subgraph/src ./src
COPY task2/booking.proto ./booking.proto
RUN npm run typecheck && npm prune --omit=dev
ENV NODE_ENV=production
USER node
EXPOSE 4001
CMD ["node", "src/index.ts"]
