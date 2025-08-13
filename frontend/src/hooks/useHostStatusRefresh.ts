import { GetStatusWithHosts } from "@bindings/easy-check/internal/services/appservice";
import { useCallback, useEffect, useState } from "react";

import { Host, HostStatusMap } from "@/types/host";
import { SeriesPoint } from "@/types/series";

export const useHostStatusRefresh = (
  hosts: Host[],
  refreshInterval: number | null,
  selectedHosts: string[],
  onDataPoint: (hostName: string, point: SeriesPoint) => void
) => {
  const [statusData, setStatusData] = useState<HostStatusMap>(new Map());

  const fetchHostsStatus = useCallback(async () => {
    if (hosts?.length === 0) {
      setStatusData(new Map());
      return;
    }

    try {
      const hostNames = hosts.map((host) => host.host);
      if (hostNames?.length === 0) {
        setStatusData(new Map());
        return;
      }

      const res = await GetStatusWithHosts(hostNames);
      const statusHosts = res?.hosts || [];
      const statusList: HostStatusMap = new Map();
      const now = Date.now();

      statusHosts.forEach((statusHost) => {
        statusList.set(statusHost.host, {
          description: statusHost.host,
          latency: statusHost.avg_latency || null,
          status: statusHost.status === "ALERT" ? "ALERT" : "RECOVERY",
          sent: false,
        });

        const hostName = statusHost.host;

        // 只为选中的主机添加趋势数据点
        if (selectedHosts.includes(hostName)) {
          const point: SeriesPoint = {
            ts: now,
            min:
              typeof statusHost.min_latency === "number"
                ? statusHost.min_latency
                : undefined,
            avg:
              typeof statusHost.avg_latency === "number"
                ? statusHost.avg_latency
                : undefined,
            max:
              typeof statusHost.max_latency === "number"
                ? statusHost.max_latency
                : undefined,
            loss:
              typeof statusHost.packet_loss === "number"
                ? statusHost.packet_loss
                : undefined,
          };

          onDataPoint(hostName, point);
          console.log(`Updated trend data for selected host: ${hostName}`);
        }
      });

      setStatusData(statusList);
      console.log(
        `Auto refresh completed at ${new Date().toLocaleTimeString()}`
      );
    } catch (err) {
      console.error("Error fetching latency data:", err);
    }
  }, [hosts, selectedHosts, onDataPoint]);

  useEffect(() => {
    if (hosts.length > 0) {
      fetchHostsStatus();
    } else {
      setStatusData(new Map());
    }

    let intervalId: NodeJS.Timeout | null = null;
    if (refreshInterval && hosts.length > 0) {
      intervalId = setInterval(fetchHostsStatus, refreshInterval);
    }

    return () => {
      if (intervalId) clearInterval(intervalId);
    };
  }, [hosts, refreshInterval, fetchHostsStatus]);

  return {
    statusData,
    fetchHostsStatus,
  };
};
