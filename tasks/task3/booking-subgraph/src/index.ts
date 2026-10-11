import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';

import { grpcBookingSource } from './bookings.ts';
import { createSchema, type Context } from './schema.ts';

const port = Number(process.env.PORT ?? 4001);
const bookingServiceAddr = process.env.BOOKING_SERVICE_ADDR ?? 'localhost:9090';
const protoPath = process.env.BOOKING_PROTO_PATH ?? 'booking.proto';
const timeoutMs = Number(process.env.BOOKING_SERVICE_TIMEOUT_MS ?? 2000);

const server = new ApolloServer<Context>({
  schema: createSchema(grpcBookingSource(bookingServiceAddr, protoPath, timeoutMs)),
});

const { url } = await startStandaloneServer(server, {
  listen: { port },
  context: async ({ req }) => {
    const header = req.headers['userid'];
    return { requesterId: Array.isArray(header) ? header[0] : header };
  },
});
console.log(JSON.stringify({ msg: 'booking subgraph ready', url, bookingServiceAddr }));
