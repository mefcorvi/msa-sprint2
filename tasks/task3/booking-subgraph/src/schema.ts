import { buildSubgraphSchema } from '@apollo/subgraph';
import { GraphQLError, parse } from 'graphql';

import type { Booking, BookingSource } from './bookings.ts';

export interface Context {
  /** Id of the requesting user, taken from the `userid` header forwarded by the gateway. */
  requesterId?: string;
}

const typeDefs = parse(`
  extend schema @link(url: "https://specs.apollo.dev/federation/v2.5", import: ["@key"])

  type Booking {
    id: ID!
    userId: String!
    hotelId: String!
    promoCode: String
    discountPercent: Float!
    price: Float!
    createdAt: String!
    hotel: Hotel!
  }

  type Hotel @key(fields: "id", resolvable: false) {
    id: ID!
  }

  type Query {
    bookingsByUser(userId: String!): [Booking!]!
  }
`);

/** ACL: a user may read only their own bookings. */
function assertCanReadBookings(requesterId: string | undefined, userId: string): void {
  if (!requesterId) {
    throw new GraphQLError('Authentication required: pass the userid header', {
      extensions: { code: 'UNAUTHENTICATED' },
    });
  }
  if (requesterId !== userId) {
    throw new GraphQLError('Access denied: a user can read only their own bookings', {
      extensions: { code: 'FORBIDDEN' },
    });
  }
}

export function createSchema(bookings: BookingSource) {
  const resolvers = {
    Query: {
      bookingsByUser: async (_: unknown, args: { userId: string }, ctx: Context) => {
        try {
          assertCanReadBookings(ctx.requesterId, args.userId);
        } catch (err) {
          console.log(JSON.stringify({ msg: 'bookings access denied', requesterId: ctx.requesterId ?? null, userId: args.userId }));
          throw err;
        }
        const result = await bookings.listByUser(args.userId);
        console.log(JSON.stringify({ msg: 'bookings returned', userId: args.userId, count: result.length }));
        return result;
      },
    },
    Booking: {
      // proto3 strings cannot be null: booking-service sends "" for a booking without a promo code.
      promoCode: (b: Booking) => b.promoCode || null,
      hotel: (b: Booking) => ({ __typename: 'Hotel', id: b.hotelId }),
    },
  };
  return buildSubgraphSchema([{ typeDefs, resolvers }]);
}
