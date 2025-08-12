import { useCallback, useMemo, useRef } from 'react';

interface DataPoint {
  ts: number;
  [key: string]: any;
}

interface OptimizationOptions {
  maxDataPoints?: number;
  enableLTTB?: boolean;
  viewportOnly?: boolean;
}

/**
 * Grafana 风格的数据优化 Hook
 * 实现数据降采样、视口优化等高级功能
 */
export function useGrafanaStyleOptimization(
  data: DataPoint[],
  xAxisDomain: [number, number],
  options: OptimizationOptions = {}
) {
  const {
    maxDataPoints = 1000,
    enableLTTB = true,
    viewportOnly = true,
  } = options;

  const animationFrameRef = useRef<number | null>(null);

  // LTTB (Largest Triangle Three Buckets) 降采样算法
  const lttbDownsample = useCallback((data: DataPoint[], threshold: number) => {
    if (data.length <= threshold) return data;
    if (threshold <= 2) return [data[0], data[data.length - 1]];

    const sampled: DataPoint[] = [];
    const bucketSize = (data.length - 2) / (threshold - 2);
    
    sampled.push(data[0]); // 保留第一个点

    for (let i = 1; i < threshold - 1; i++) {
      const bucketStart = Math.floor(i * bucketSize) + 1;
      const bucketEnd = Math.floor((i + 1) * bucketSize) + 1;
      const bucketCenter = (bucketStart + bucketEnd) / 2;

      let maxArea = -1;
      let maxAreaIndex = bucketStart;

      // 计算三角形面积找到最重要的点
      for (let j = bucketStart; j < bucketEnd; j++) {
        const area = Math.abs(
          (data[Math.floor(bucketCenter)].ts - data[bucketStart - 1].ts) *
          (data[j].ts - data[bucketStart - 1].ts) -
          (data[Math.floor(bucketCenter)].ts - data[bucketStart - 1].ts) *
          (data[j].ts - data[Math.floor(bucketCenter)].ts)
        );
        
        if (area > maxArea) {
          maxArea = area;
          maxAreaIndex = j;
        }
      }
      
      sampled.push(data[maxAreaIndex]);
    }

    sampled.push(data[data.length - 1]); // 保留最后一个点
    return sampled;
  }, []);

  // 视口数据过滤
  const viewportFilter = useCallback((data: DataPoint[]) => {
    if (!viewportOnly) return data;
    
    const [start, end] = xAxisDomain;
    const buffer = (end - start) * 0.1; // 10% buffer
    
    return data.filter(point => 
      point.ts >= start - buffer && point.ts <= end + buffer
    );
  }, [xAxisDomain, viewportOnly]);

  // 优化后的数据
  const optimizedData = useMemo(() => {
    let processedData = [...data];
    
    // 1. 先进行视口过滤
    processedData = viewportFilter(processedData);
    
    // 2. 如果启用 LTTB 且数据点超过阈值，进行降采样
    if (enableLTTB && processedData.length > maxDataPoints) {
      processedData = lttbDownsample(processedData, maxDataPoints);
    }
    
    return processedData;
  }, [data, viewportFilter, enableLTTB, maxDataPoints, lttbDownsample]);

  // Grafana 风格的平滑鼠标跟踪
  const smoothMouseTracking = useCallback((callback: () => void) => {
    if (animationFrameRef.current) {
      cancelAnimationFrame(animationFrameRef.current);
    }
    
    animationFrameRef.current = requestAnimationFrame(callback);
  }, []);

  return {
    optimizedData,
    smoothMouseTracking,
    dataReduction: data.length - optimizedData.length,
    compressionRatio: optimizedData.length / data.length,
  };
}

export default useGrafanaStyleOptimization;
