import { useState, useCallback } from "react";
import { GetHistoryWithHosts } from "@bindings/easy-check/internal/services/appservice";
import { HostSeriesMap, SeriesPoint } from "@/types/series";

// 配置常量
const PING_INTERVAL_SECONDS = 30; // 从配置文件中的 ping.interval，每30秒产生一个数据点
const DEFAULT_TIME_WINDOW_MINUTES = 60; // 默认保留1小时的数据

// 根据时间窗口计算该时间范围内应该有多少个数据点
const calculateExpectedDataPoints = (timeWindowMinutes: number) => {
  // 每分钟的数据点数 = 60 / PING_INTERVAL_SECONDS
  const pointsPerMinute = 60 / PING_INTERVAL_SECONDS;
  return Math.ceil(pointsPerMinute * timeWindowMinutes);
};

// 计算滑动窗口的最大数据点数（包含缓冲区）
const calculateMaxDataPoints = (
  timeWindowMinutes: number = DEFAULT_TIME_WINDOW_MINUTES
) => {
  const expectedPoints = calculateExpectedDataPoints(timeWindowMinutes);
  // 为避免频繁删除，增加20%的缓冲区
  return Math.ceil(expectedPoints * 1.2);
};

export const useHistoryData = () => {
  const [historyMap, setHistoryMap] = useState<HostSeriesMap>({});

  // 添加新数据点的函数（滑动窗口机制）
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

      // 实现滑动窗口：根据时间窗口动态计算最大数据点数
      const maxPoints = calculateMaxDataPoints();
      let finalData = newData;
      if (finalData.length > maxPoints) {
        // 计算需要保留的数据点数（保留90%，留出缓冲区）
        const keepCount = Math.floor(maxPoints * 0.9);
        const removedCount = finalData.length - keepCount;
        finalData = finalData.slice(-keepCount);

        console.log(
          `Sliding window applied for ${hostName}: removed ${removedCount} old points, kept ${keepCount} points`
        );
      }

      return {
        ...prev,
        [hostName]: finalData,
      };
    });
  }, []);

  // 为新选中的主机加载历史数据
  const loadHistoryForHost = useCallback(
    async (hostName: string, timeRangeMinutes = 30) => {
      try {
        const now = Date.now();
        const startTime = now - timeRangeMinutes * 60 * 1000;
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

            // 应用滑动窗口限制 - 根据实际的时间范围计算最大数据点数
            const maxPoints = calculateMaxDataPoints(timeRangeMinutes);
            let finalData = historicalPoints;
            if (finalData.length > maxPoints) {
              finalData = finalData.slice(-maxPoints);
            }

            setHistoryMap((prev) => ({
              ...prev,
              [hostName]: finalData,
            }));

            console.log(
              `Loaded ${
                finalData.length
              } history points for host: ${hostName} (expected ~${calculateExpectedDataPoints(
                timeRangeMinutes
              )} for ${timeRangeMinutes}min window)`
            );
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
      const { [hostName]: removed, ...rest } = prev;
      return rest;
    });
  }, []);

  return {
    historyMap,
    addDataPoint,
    loadHistoryForHost,
    clearHistoryForHost,
  };
};
