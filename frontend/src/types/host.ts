export interface Host {
  host: string;
  description: string;
}

export interface HostStatus {
  id: string;
  name: string;
  latency: number; // 对应 tsdb 中的 avg_latency
  status?: "ALERT" | "RECOVERY";
  sent?: boolean;
}
