import { useState, useCallback } from "react";
import { toaster } from "@/components/ui/toaster";
import { HostStatusMap } from "@/types/host";

export const useHostSelection = (
  initialStatusData: HostStatusMap,
  onHostSelected: (hostName: string) => void,
  onHostDeselected: (hostName: string) => void
) => {
  const [selectedHosts, setSelectedHosts] = useState<string[]>([]);
  const [statusData, setStatusData] =
    useState<HostStatusMap>(initialStatusData);

  const updateStatusData = useCallback((newStatusData: HostStatusMap) => {
    setStatusData(newStatusData);
  }, []);

  const handleHostSelection = useCallback(
    (host: string) => {
      const s = statusData.get(host);
      if (!s?.latency) {
        toaster.create({
          title: "暂无监控数据",
          description: s?.status === "ALERT" ? "主机告警中" : "暂无数据",
          type: "warning",
        });
        return;
      }

      setSelectedHosts((prev) => {
        const exists = prev.includes(host);
        let next = exists ? prev.filter((h) => h !== host) : [...prev, host];

        if (next.length > 5) {
          toaster.create({
            title: "最多选择5个主机",
            description: "已达到选择上限",
            type: "warning",
          });
          next = next.slice(0, 5);
        }

        // 通知外部组件主机选择状态变化
        if (!exists && next.includes(host)) {
          onHostSelected(host);
        } else if (exists && !next.includes(host)) {
          onHostDeselected(host);
        }

        return next;
      });
    },
    [statusData, onHostSelected, onHostDeselected]
  );

  const clearSelection = useCallback(() => {
    selectedHosts.forEach((host) => onHostDeselected(host));
    setSelectedHosts([]);
  }, [selectedHosts, onHostDeselected]);

  return {
    selectedHosts,
    handleHostSelection,
    clearSelection,
    updateStatusData,
  };
};
