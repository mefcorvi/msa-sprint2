export interface Hotel {
  id: string;
  /** The monolith has no hotel name, so its description is used. */
  name: string | null;
  city: string | null;
  rating: number | null;
  description: string | null;
  operational: boolean;
  fullyBooked: boolean;
}

export interface HotelSource {
  findById(id: string): Promise<Hotel | null>;
}

interface MonolithHotel {
  id: string;
  city: string | null;
  rating: number | null;
  description: string | null;
  operational: boolean;
  fullyBooked: boolean;
}

export function fromMonolith(h: MonolithHotel): Hotel {
  return {
    id: h.id,
    name: h.description,
    city: h.city,
    rating: h.rating,
    description: h.description,
    operational: h.operational,
    fullyBooked: h.fullyBooked,
  };
}

/** Reads hotels from the monolith REST API until hotel-service is extracted. */
export function monolithHotelSource(baseUrl: string, timeoutMs: number): HotelSource {
  return {
    async findById(id) {
      const response = await fetch(`${baseUrl}/api/hotels/${encodeURIComponent(id)}`, {
        signal: AbortSignal.timeout(timeoutMs),
      });
      if (response.status === 404) {
        return null;
      }
      if (!response.ok) {
        throw new Error(`monolith GET /api/hotels/${id}: HTTP ${response.status}`);
      }
      return fromMonolith((await response.json()) as MonolithHotel);
    },
  };
}
