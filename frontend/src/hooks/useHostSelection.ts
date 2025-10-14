import { useCallback, useState } from "react";

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

        if (exists) {
          // 取消选择
          const next = prev.filter((h) => h !== host);
          onHostDeselected(host);
          return next;
        } else {
          // 添加选择
          if (prev.length >= 5) {
            toaster.create({
              title: "最多选择5个主机",
              description: "已达到选择上限",
              type: "warning",
            });
            return prev; // 不添加新主机，保持原有选择
          }

          const next = [...prev, host];
          onHostSelected(host);
          return next;
        }
      });
    },
    [statusData, onHostSelected, onHostDeselected]
  );

  const clearSelection = useCallback(() => {
    setSelectedHosts((prev) => {
      prev.forEach((host) => onHostDeselected(host));
      return [];
    });
  }, [onHostDeselected]);

  return {
    selectedHosts,
    handleHostSelection,
    clearSelection,
    updateStatusData,
  };
};
