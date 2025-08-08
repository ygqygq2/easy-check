export interface SeriesPoint {
  ts: number; // epoch ms
  min?: number;
  avg?: number;
  max?: number;
  loss?: number; // percent 0-100
}

export type HostSeriesMap = Record<string, SeriesPoint[]>; // host -> points
