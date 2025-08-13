import { GetHistoryWithHosts } from "@bindings/easy-check/internal/services/appservice";
import { useCallback, useState } from "react";

import { HostSeriesMap, SeriesPoint } from "@/types/series";
// 已采用服务端步长与范围控制，移除旧的客户端点数/步长计算逻辑。

export const useHistoryData = () => {
  const [historyMap, setHistoryMap] = useState<HostSeriesMap>({});
  const [lastStepSeconds, setLastStepSeconds] = useState<number | null>(null);

  // 添加新数据点：不再强制 48h 裁剪，保留全部已加载历史（后续可引入 LRU 或分段缓存）
  const addDataPoint = useCallback((hostName: string, point: SeriesPoint) => {
    setHistoryMap((prev) => {
      const existingData = prev[hostName] || [];
      const newData = [...existingData];

      // 检查是否是重复的时间戳（避免重复添加）
      if (newData.length === 0 || newData[newData.length - 1].ts !== point.ts) {
        newData.push(point);
      } else {
        // 更新最后一个点的数据
        newData[newData.length - 1] = point;
      }

      return {
        ...prev,
        [hostName]: newData,
      };
    });
  }, []);

  // 检测缓存中的数据缺失并从数据库补全
  const fillMissingData = useCallback(
    async (hostName: string, timeRangeMinutes = 30) => {
      const existingData = historyMap[hostName] || [];

      if (existingData.length === 0) {
        // 如果完全没有数据，直接加载历史数据
        return loadHistoryForHost(hostName, timeRangeMinutes);
      }

      const now = Date.now();
      const startTime = now - timeRangeMinutes * 60 * 1000;

      // 找出缺失的时间段
      const missingRanges: Array<{ start: number; end: number }> = [];

      // 检查从开始时间到第一个数据点之间是否有缺失
      const firstDataPoint = existingData[0];
      if (firstDataPoint.ts > startTime + 2 * 60 * 1000) {
        // 超过2分钟的间隔
        missingRanges.push({
          start: startTime,
          end: firstDataPoint.ts - 60 * 1000, // 留1分钟缓冲
        });
      }

      // 检查数据点之间的间隔
      for (let i = 1; i < existingData.length; i++) {
        const prevPoint = existingData[i - 1];
        const currentPoint = existingData[i];
        const gap = currentPoint.ts - prevPoint.ts;

        // 如果间隔超过3分钟，认为有数据缺失
        if (gap > 3 * 60 * 1000) {
          missingRanges.push({
            start: prevPoint.ts + 60 * 1000,
            end: currentPoint.ts - 60 * 1000,
          });
        }
      }

      // 检查最后一个数据点到现在是否有缺失
      const lastDataPoint = existingData[existingData.length - 1];
      if (now - lastDataPoint.ts > 2 * 60 * 1000) {
        missingRanges.push({
          start: lastDataPoint.ts + 60 * 1000,
          end: now,
        });
      }

      // 为每个缺失的时间段补全数据
      for (const range of missingRanges) {
        try {
          await loadHistoryForHost(
            hostName,
            Math.ceil((range.end - range.start) / (60 * 1000)), // 转换为分钟
            range.start,
            range.end
          );
        } catch (error) {
          console.error(`Failed to fill missing data for ${hostName}:`, error);
        }
      }
    },
    [historyMap]
  );

  // 智能加载历史数据，支持指定时间范围
  const loadHistoryForHost = useCallback(
    async (
      hostName: string,
      timeRangeMinutes = 30,
      customStartTime?: number,
      customEndTime?: number
    ) => {
      try {
        const now = customEndTime || Date.now();
        const startTime = customStartTime || now - timeRangeMinutes * 60 * 1000;
        const endTime = now;

        const historyRes = await GetHistoryWithHosts(
          [hostName],
          startTime,
          endTime,
          0 // 让服务端决定分辨率
        );

        if (historyRes?.step_seconds) {
          setLastStepSeconds(historyRes.step_seconds);
        }
        if (historyRes?.hosts && historyRes.hosts.length > 0) {
          const hostData = historyRes.hosts[0];
          const historicalPoints: SeriesPoint[] = [];

          if (hostData.series) {
            const avgData = hostData.series["avg_latency"] || [];
            const minData = hostData.series["min_latency"] || [];
            const maxData = hostData.series["max_latency"] || [];
            const lossData = hostData.series["packet_loss"] || [];

            // 创建时间戳到数据点的映射
            const pointsMap: { [ts: number]: SeriesPoint } = {};

            avgData.forEach((point) => {
              if (!pointsMap[point.timestamp]) {
                pointsMap[point.timestamp] = { ts: point.timestamp };
              }
              pointsMap[point.timestamp].avg = point.value;
            });

            minData.forEach((point) => {
              if (!pointsMap[point.timestamp]) {
                pointsMap[point.timestamp] = { ts: point.timestamp };
              }
              pointsMap[point.timestamp].min = point.value;
            });

            maxData.forEach((point) => {
              if (!pointsMap[point.timestamp]) {
                pointsMap[point.timestamp] = { ts: point.timestamp };
              }
              pointsMap[point.timestamp].max = point.value;
            });

            lossData.forEach((point) => {
              if (!pointsMap[point.timestamp]) {
                pointsMap[point.timestamp] = { ts: point.timestamp };
              }
              pointsMap[point.timestamp].loss = point.value;
            });

            // 转换为数组并排序
            historicalPoints.push(
              ...Object.values(pointsMap).sort((a, b) => a.ts - b.ts)
            );

            if (historicalPoints.length > 0) {
              setHistoryMap((prev) => {
                const existingData = prev[hostName] || [];

                // 合并新数据和现有数据，去重并排序
                const allPoints = [...existingData, ...historicalPoints];
                const uniquePoints = Array.from(
                  new Map(allPoints.map((point) => [point.ts, point])).values()
                ).sort((a, b) => a.ts - b.ts);

                return {
                  ...prev,
                  [hostName]: uniquePoints,
                };
              });
            }
          }
        }
      } catch (err) {
        console.error(`Error loading history data for host ${hostName}:`, err);
      }
    },
    []
  );

  // 清除主机的历史数据
  const clearHistoryForHost = useCallback((hostName: string) => {
    setHistoryMap((prev) => {
      const { [hostName]: _removed, ...rest } = prev;
      return rest;
    });
  }, []);

  return {
    historyMap,
    lastStepSeconds,
    addDataPoint,
    loadHistoryForHost,
    fillMissingData,
    clearHistoryForHost,
  };
};
