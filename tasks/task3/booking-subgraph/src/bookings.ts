import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';

export interface Booking {
  id: string;
  userId: string;
  hotelId: string;
  promoCode: string;
  discountPercent: number;
  price: number;
  createdAt: string;
}

export interface BookingSource {
  listByUser(userId: string): Promise<Booking[]>;
}

type ListBookings = (
  request: { userId: string },
  options: { deadline: Date },
  callback: (err: grpc.ServiceError | null, response?: { bookings: Booking[] }) => void,
) => void;

/** Reads bookings from booking-service over gRPC (the shared booking.proto contract). */
export function grpcBookingSource(address: string, protoPath: string, timeoutMs: number): BookingSource {
  const definition = protoLoader.loadSync(protoPath, { longs: String, defaults: true });
  const proto = grpc.loadPackageDefinition(definition) as any;
  const client = new proto.booking.BookingService(address, grpc.credentials.createInsecure());
  const listBookings: ListBookings = client.listBookings.bind(client);

  return {
    listByUser: (userId) =>
      new Promise((resolve, reject) => {
        listBookings({ userId }, { deadline: new Date(Date.now() + timeoutMs) }, (err, response) =>
          err ? reject(err) : resolve(response?.bookings ?? []),
        );
      }),
  };
}
