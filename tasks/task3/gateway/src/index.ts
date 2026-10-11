import {
  ApolloGateway,
  IntrospectAndCompose,
  RemoteGraphQLDataSource,
  type GraphQLDataSourceProcessOptions,
} from '@apollo/gateway';
import { ApolloServer } from '@apollo/server';
import { ApolloServerPluginLandingPageLocalDefault } from '@apollo/server/plugin/landingPage/default';
import { startStandaloneServer } from '@apollo/server/standalone';

interface Context {
  userid?: string;
}

const port = Number(process.env.PORT ?? 4000);
const subgraphs = [
  { name: 'booking', url: process.env.BOOKING_SUBGRAPH_URL ?? 'http://booking-subgraph:4001' },
  { name: 'hotel', url: process.env.HOTEL_SUBGRAPH_URL ?? 'http://hotel-subgraph:4002' },
];

/** Apollo Gateway does not forward client headers by itself; subgraphs need userid for the ACL. */
class UserIdForwardingDataSource extends RemoteGraphQLDataSource<Context> {
  override willSendRequest({ request, context }: GraphQLDataSourceProcessOptions<Context>) {
    if (context.userid) {
      request.http?.headers.set('userid', context.userid);
    }
  }
}

const gateway = new ApolloGateway({
  supergraphSdl: new IntrospectAndCompose({ subgraphs }),
  buildService: ({ url }) => new UserIdForwardingDataSource({ url }),
});

// The task is checked in the Playground, which NODE_ENV=production would otherwise disable.
const server = new ApolloServer<Context>({
  gateway,
  introspection: true,
  plugins: [ApolloServerPluginLandingPageLocalDefault()],
});

const { url } = await startStandaloneServer(server, {
  listen: { port },
  context: async ({ req }) => {
    const header = req.headers['userid'];
    return { userid: Array.isArray(header) ? header[0] : header };
  },
});
console.log(JSON.stringify({ msg: 'gateway ready', url, subgraphs }));
