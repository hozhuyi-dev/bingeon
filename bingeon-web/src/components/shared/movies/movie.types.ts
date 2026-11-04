/** A movie as a card shows it: what every list in either API gives. */
export interface MovieSummary {
  id: number;
  title: string;
  year?: number;
  poster?: string;
  rating?: number;
}
