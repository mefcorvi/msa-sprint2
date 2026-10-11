import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';

import { monolithHotelSource } from './hotels.ts';
import { createSchema } from './schema.ts';

const port = Number(process.env.PORT ?? 4002);
const monolithUrl = process.env.MONOLITH_URL ?? 'http://localhost:8084';
const timeoutMs = Number(process.env.MONOLITH_TIMEOUT_MS ?? 2000);

const server = new ApolloServer({ schema: createSchema(monolithHotelSource(monolithUrl, timeoutMs)) });

const { url } = await startStandaloneServer(server, { listen: { port } });
console.log(JSON.stringify({ msg: 'hotel subgraph ready', url, monolithUrl }));
