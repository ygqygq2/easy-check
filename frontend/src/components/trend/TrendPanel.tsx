import { useMemo } from "react";
import { Box, Grid, GridItem, HStack, Text } from "@chakra-ui/react";
import PingLatencyChart from "./PingLatencyChart";
import PacketLossChart from "./PacketLossChart";
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
          if (typeof p.min === "number") row[`${h}:min`] = p.min;
          if (typeof p.avg === "number") row[`${h}:avg`] = p.avg;
          if (typeof p.max === "number") row[`${h}:max`] = p.max;
          if (typeof p.loss === "number") row[`${h}:loss`] = p.loss;
          if (typeof p.min === "number" && typeof p.max === "number") {
            const range = p.max - p.min;
            if (range >= 0) row[`${h}:range`] = range;
          }
        }
      });
      return row;
    });
  }, [seriesMap, hosts]);
}

export default function TrendPanel({ selectedHosts, seriesMap }: Props) {
  const merged = useMerged(seriesMap, selectedHosts);
  const colorMap = useMemo(() => {
    const m: Record<string, string> = {};
    selectedHosts.forEach((h, i) => (m[h] = COLORS[i % COLORS.length]));
    return m;
  }, [selectedHosts]);

  return (
    <Box w="100%" h="100%" display="flex" flexDirection="column">
      <HStack gap="4" px="2" py="2">
        <Text>历史趋势（最近10分钟）</Text>
        <Text fontSize="sm" color="gray.500">
          已选 {selectedHosts.length} 台（最多 5 台）
        </Text>
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
      ) : (
        <Grid templateRows="2fr 1fr" h="100%" gap="2">
          <GridItem minH={{ base: 160, md: 220 }}>
            <PingLatencyChart
              data={merged}
              selectedHosts={selectedHosts}
              colorMap={colorMap}
            />
          </GridItem>
          <GridItem minH={{ base: 120, md: 140 }}>
            <PacketLossChart
              data={merged}
              selectedHosts={selectedHosts}
              colorMap={colorMap}
            />
          </GridItem>
        </Grid>
      )}
    </Box>
  );
}
