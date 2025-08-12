import { useState, useCallback, useRef, useMemo } from "react";

interface TooltipState {
  x: number;
  y: number;
  timestamp: number;
  visible: boolean;
}

interface UseTooltipOptimizationOptions {
  data: any[];
  selectedHosts: string[];
  xAxisDomain: [number, number];
  debounceMs?: number;
}

/**
 * 优化图表 Tooltip 性能的自定义 Hook
 * 包含防抖、数据查找优化和性能监控
 */
export function useTooltipOptimization({
  data,
  selectedHosts,
  xAxisDomain,
  debounceMs = 16, // 约 60fps
}: UseTooltipOptimizationOptions) {
  const [tooltipState, setTooltipState] = useState<TooltipState | null>(null);
  const debounceTimer = useRef<number | null>(null);
  const lastUpdate = useRef<number>(0);

  // 使用 useMemo 缓存排序后的数据以提高查找性能
  const sortedData = useMemo(() => {
    if (!data || data.length === 0) return [];
    return [...data].sort((a: any, b: any) => a.ts - b.ts);
  }, [data]);

  // 优化的二分查找数据点
  const findOptimalDataPoint = useCallback(
    (mouseX: number, chartWidth: number) => {
      if (!sortedData || sortedData.length === 0) return null;

      const [start, end] = xAxisDomain;
      const timespan = end - start;
      const ratio = mouseX / chartWidth;
      const targetTimestamp = start + ratio * timespan;

      // 二分查找
      let left = 0;
      let right = sortedData.length - 1;
      let bestMatch = sortedData[0];
      let bestDiff = Math.abs(sortedData[0].ts - targetTimestamp);

      while (left <= right) {
        const mid = Math.floor((left + right) / 2);
        const currentPoint = sortedData[mid];
        const diff = Math.abs(currentPoint.ts - targetTimestamp);

        if (diff < bestDiff) {
          bestDiff = diff;
          bestMatch = currentPoint;
        }

        if (currentPoint.ts < targetTimestamp) {
          left = mid + 1;
        } else {
          right = mid - 1;
        }
      }

      // 检查最近的几个点以确保最优匹配
      const candidates = [left - 1, left, right, right + 1]
        .filter((idx) => idx >= 0 && idx < sortedData.length)
        .map((idx) => sortedData[idx]);

      for (const candidate of candidates) {
        const diff = Math.abs(candidate.ts - targetTimestamp);
        if (diff < bestDiff) {
          bestDiff = diff;
          bestMatch = candidate;
        }
      }

      // 验证是否有选中主机的数据
      const hasData = selectedHosts.some(
        (host) => typeof bestMatch[`${host}:avg`] === "number"
      );

      return hasData ? bestMatch : null;
    },
    [sortedData, xAxisDomain, selectedHosts]
  );

  // 防抖的鼠标移动处理器
  const handleMouseMove = useCallback(
    (mouseX: number, mouseY: number, chartWidth: number) => {
      const now = Date.now();

      // 防抖处理
      if (debounceTimer.current) {
        window.clearTimeout(debounceTimer.current);
      }

      debounceTimer.current = window.setTimeout(() => {
        const nearestPoint = findOptimalDataPoint(mouseX, chartWidth);

        if (nearestPoint) {
          setTooltipState({
            x: mouseX,
            y: mouseY,
            timestamp: nearestPoint.ts,
            visible: true,
          });
        } else {
          setTooltipState(null);
        }

        lastUpdate.current = now;
      }, debounceMs);
    },
    [findOptimalDataPoint, debounceMs]
  );

  const handleMouseLeave = useCallback(() => {
    if (debounceTimer.current) {
      window.clearTimeout(debounceTimer.current);
    }
    setTooltipState(null);
  }, []);

  const resetTooltip = useCallback(() => {
    if (debounceTimer.current) {
      window.clearTimeout(debounceTimer.current);
    }
    setTooltipState(null);
  }, []);

  return {
    tooltipState,
    handleMouseMove,
    handleMouseLeave,
    resetTooltip,
    // 性能指标
    getPerformanceMetrics: () => ({
      lastUpdateTime: lastUpdate.current,
      dataPointsCount: sortedData.length,
      selectedHostsCount: selectedHosts.length,
    }),
  };
}

export default useTooltipOptimization;
