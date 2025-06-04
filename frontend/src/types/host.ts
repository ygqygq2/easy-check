export interface Host {
  host: string;
  description: string;
}

export interface HostStatus {
  description: string;
  latency: number | null; // 对应 tsdb 中的 avg_latency
  status?: "ALERT" | "RECOVERY";
  sent?: boolean;
}

export type HostStatusMap = Map<string, HostStatus>;
