import { useState, useCallback } from "react";
import { GetHistoryWithHosts } from "@bindings/easy-check/internal/services/appservice";
import { HostSeriesMap, SeriesPoint } from "@/types/series";
import { useConfig } from "./useConfig";

// 根据时间窗口计算该时间范围内应该有多少个数据点
const calculateExpectedDataPoints = (
  timeWindowMinutes: number,
  intervalSeconds: number
) => {
  const pointsPerMinute = 60 / intervalSeconds;
  return Math.ceil(pointsPerMinute * timeWindowMinutes);
};

// 计算滑动窗口的最大数据点数（包含缓冲区）
const calculateMaxDataPoints = (
  timeWindowMinutes: number,
  intervalSeconds: number
) => {
  const expectedPoints = calculateExpectedDataPoints(
    timeWindowMinutes,
    intervalSeconds
  );
  // 为避免频繁删除，增加20%的缓冲区
  return Math.ceil(expectedPoints * 1.2);
};

export const useHistoryData = () => {
  const [historyMap, setHistoryMap] = useState<HostSeriesMap>({});
  const { getPingInterval } = useConfig();

  // 添加新数据点的函数（滑动窗口机制）
  const addDataPoint = useCallback(
    (hostName: string, point: SeriesPoint) => {
      setHistoryMap((prev) => {
        const existingData = prev[hostName] || [];
        const newData = [...existingData];

        // 检查是否是重复的时间戳（避免重复添加）
        if (
          newData.length === 0 ||
          newData[newData.length - 1].ts !== point.ts
        ) {
          newData.push(point);
        } else {
          // 更新最后一个点的数据
          newData[newData.length - 1] = point;
        }

        // 实现滑动窗口：根据时间窗口动态计算最大数据点数
        // 使用60分钟作为默认窗口，从配置获取数据间隔
        const intervalSeconds = getPingInterval();
        const maxPoints = calculateMaxDataPoints(60, intervalSeconds);
        let finalData = newData;
        if (finalData.length > maxPoints) {
          // 计算需要保留的数据点数（保留90%，留出缓冲区）
          const keepCount = Math.floor(maxPoints * 0.9);
          const removedCount = finalData.length - keepCount;
          finalData = finalData.slice(-keepCount);

          console.log(
            `Sliding window applied for ${hostName}: removed ${removedCount} old points, kept ${keepCount} points (interval: ${intervalSeconds}s)`
          );
        }

        return {
          ...prev,
          [hostName]: finalData,
        };
      });
    },
    [getPingInterval]
  );

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
          console.log(
            `Filling missing data for ${hostName} from ${new Date(
              range.start
            ).toLocaleTimeString()} to ${new Date(
              range.end
            ).toLocaleTimeString()}`
          );

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

        // 根据时间范围计算步长（秒）
        let step = 60; // 默认1分钟
        if (timeRangeMinutes <= 60) step = 60;
        else if (timeRangeMinutes <= 360) step = 300;
        else if (timeRangeMinutes <= 1440) step = 900;
        else if (timeRangeMinutes <= 10080) step = 3600;
        else step = 7200;

        const historyRes = await GetHistoryWithHosts(
          [hostName],
          startTime,
          endTime,
          step
        );

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

                // 应用滑动窗口限制
                const intervalSeconds = getPingInterval();
                const maxPoints = calculateMaxDataPoints(
                  timeRangeMinutes,
                  intervalSeconds
                );
                let finalData = uniquePoints;
                if (finalData.length > maxPoints) {
                  finalData = finalData.slice(-maxPoints);
                }

                return {
                  ...prev,
                  [hostName]: finalData,
                };
              });

              console.log(
                `Loaded/merged ${historicalPoints.length} history points for host: ${hostName}`
              );
            }
          }
        }
      } catch (err) {
        console.error(`Error loading history data for host ${hostName}:`, err);
      }
    },
    [getPingInterval]
  );

  // 清除主机的历史数据
  const clearHistoryForHost = useCallback((hostName: string) => {
    setHistoryMap((prev) => {
      const { [hostName]: removed, ...rest } = prev;
      return rest;
    });
  }, []);

  return {
    historyMap,
    addDataPoint,
    loadHistoryForHost,
    fillMissingData,
    clearHistoryForHost,
  };
};
