import { useMemo, useState } from "react";
import { Box, HStack, Text } from "@chakra-ui/react";
import PingLatencyChart from "./PingLatencyChart";
import TimeRangeSelector, { TimeRange, TIME_RANGES } from "./TimeRangeSelector";
import { HostSeriesMap } from "@/types/series";

const COLORS = ["#3182ce", "#38a169", "#d69e2e", "#e53e3e", "#805ad5"];

interface Props {
  selectedHosts: string[];
  seriesMap: HostSeriesMap; // host -> [{ ts, min, avg, max, loss }]
}

// 将 host->points 的结构合并为 Recharts 友好的一维数组
// 每个点包含 ts 和 `${host}:min|avg|max|loss` 等字段
function useMerged(seriesMap: HostSeriesMap, hosts: string[]) {
  return useMemo(() => {
    const tsSet = new Set<number>();
    hosts.forEach((h) => (seriesMap[h] || []).forEach((p) => tsSet.add(p.ts)));
    const allTs = Array.from(tsSet).sort((a, b) => a - b);
    return allTs.map((ts) => {
      const row: any = { ts };
      hosts.forEach((h) => {
        const p = (seriesMap[h] || []).find((x) => x.ts === ts);
        if (p) {
          if (typeof p.avg === "number") row[`${h}:avg`] = p.avg;
          if (typeof p.loss === "number") row[`${h}:loss`] = p.loss;
          // 为每个主机添加延迟范围数据 [min, max]
          if (typeof p.min === "number" && typeof p.max === "number") {
            row[`${h}:range`] = [p.min, p.max];
          }
        }
      });
      return row;
    });
  }, [seriesMap, hosts]);
}

export default function TrendPanel({ selectedHosts, seriesMap }: Props) {
  const [selectedTimeRange, setSelectedTimeRange] = useState<TimeRange>(
    TIME_RANGES[0]
  ); // 默认最近10分钟

  // 根据选择的时间范围过滤和采样数据
  const filteredSeriesMap = useMemo(() => {
    const cutoff = Date.now() - selectedTimeRange.minutes * 60 * 1000;
    const filtered: HostSeriesMap = {};

    // 根据时间范围确定最大数据点数，避免性能问题
    const getMaxDataPoints = (minutes: number) => {
      if (minutes <= 60) return 200; // 1小时内：200个点
      if (minutes <= 720) return 300; // 12小时内：300个点
      if (minutes <= 1440) return 400; // 24小时内：400个点
      if (minutes <= 10080) return 500; // 7天内：500个点
      return 600; // 更长时间：600个点
    };

    const maxPoints = getMaxDataPoints(selectedTimeRange.minutes);

    Object.keys(seriesMap).forEach((host) => {
      const hostData = (seriesMap[host] || []).filter(
        (point) => point.ts >= cutoff
      );

      // 如果数据点过多，进行采样
      if (hostData.length <= maxPoints) {
        filtered[host] = hostData;
      } else {
        // 均匀采样算法
        const sampledData = [];
        const step = hostData.length / maxPoints;
        for (let i = 0; i < maxPoints; i++) {
          const index = Math.floor(i * step);
          sampledData.push(hostData[index]);
        }
        filtered[host] = sampledData;
      }
    });

    return filtered;
  }, [seriesMap, selectedTimeRange]);

  const merged = useMerged(filteredSeriesMap, selectedHosts);
  const colorMap = useMemo(() => {
    const m: Record<string, string> = {};
    selectedHosts.forEach((h, i) => (m[h] = COLORS[i % COLORS.length]));
    return m;
  }, [selectedHosts]);
  const hasData = merged.length > 0;

  return (
    <Box w="100%" h="100%" display="flex" flexDirection="column">
      <HStack gap="4" px="2" py="1" justify="space-between">
        <HStack gap="4">
          <Text>历史趋势</Text>
          <Text fontSize="sm" color="gray.500">
            已选 {selectedHosts.length} 台（最多 5 台）
          </Text>
        </HStack>
        <TimeRangeSelector
          selectedRange={selectedTimeRange}
          onRangeChange={setSelectedTimeRange}
        />
      </HStack>
      {selectedHosts.length === 0 ? (
        <Box
          flex="1"
          display="flex"
          alignItems="center"
          justifyContent="center"
          color="gray.500"
        >
          请选择主机以查看历史趋势
        </Box>
      ) : !hasData ? (
        <Box
          flex="1"
          display="flex"
          alignItems="center"
          justifyContent="center"
          color="gray.500"
        >
          暂无数据，等待刷新...
        </Box>
      ) : (
        <Box flex="1" minH={0}>
          <PingLatencyChart
            data={merged}
            selectedHosts={selectedHosts}
            colorMap={colorMap}
            timeRangeMinutes={selectedTimeRange.minutes}
          />
        </Box>
      )}
    </Box>
  );
}
