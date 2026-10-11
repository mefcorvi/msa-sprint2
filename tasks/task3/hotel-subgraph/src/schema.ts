import { buildSubgraphSchema } from '@apollo/subgraph';
import { parse } from 'graphql';

import type { Hotel, HotelSource } from './hotels.ts';

const typeDefs = parse(`
  extend schema @link(url: "https://specs.apollo.dev/federation/v2.5", import: ["@key"])

  type Hotel @key(fields: "id") {
    id: ID!
    name: String
    city: String
    rating: Float
    description: String
    operational: Boolean!
    fullyBooked: Boolean!
  }

  type Query {
    hotelsByIds(ids: [ID!]!): [Hotel!]!
  }
`);

export function createSchema(hotels: HotelSource) {
  const resolvers = {
    Hotel: {
      __resolveReference: (ref: { id: string }) => hotels.findById(ref.id),
    },
    Query: {
      hotelsByIds: async (_: unknown, args: { ids: string[] }) => {
        const found = await Promise.all(args.ids.map((id) => hotels.findById(id)));
        return found.filter((h): h is Hotel => h !== null);
      },
    },
  };
  return buildSubgraphSchema([{ typeDefs, resolvers }]);
}
