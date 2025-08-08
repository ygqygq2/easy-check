import { useMemo } from "react";
import { HStack, Text, Box } from "@chakra-ui/react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip as RTooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";

type SeriesPoint = { ts: number; [host: string]: number };

interface TrendModalProps {
  selectedHosts: string[];
  // 数据格式：每个 host -> [{ ts, value }]
  dataMap: Record<string, Array<{ ts: number; value: number }>>;
}

const COLORS = ["#3182ce", "#38a169", "#d69e2e", "#e53e3e", "#805ad5"];

export function TrendModal({ selectedHosts, dataMap }: TrendModalProps) {
  // 将各 host 的序列按时间戳对齐
  const mergedData: SeriesPoint[] = useMemo(() => {
    const tsSet = new Set<number>();
    selectedHosts.forEach((h) => {
      (dataMap[h] || []).forEach((p) => tsSet.add(p.ts));
    });
    const allTs = Array.from(tsSet).sort((a, b) => a - b);
    return allTs.map((ts) => {
      const entry: SeriesPoint = { ts } as any;
      selectedHosts.forEach((h) => {
        const v = (dataMap[h] || []).find((p) => p.ts === ts)?.value;
        if (typeof v === "number") entry[h] = v;
      });
      return entry;
    });
  }, [selectedHosts, dataMap]);

  // X 轴时间格式化
  const formatTime = (ts: number) => {
    const d = new Date(ts);
    return `${d.getHours().toString().padStart(2, "0")}:${d
      .getMinutes()
      .toString()
      .padStart(2, "0")}`;
  };

  return (
    <Box w="100%" h="100%" display="flex" flexDirection="column">
      <HStack gap="4" px="2" py="2">
        <Text>历史趋势（最近10分钟）</Text>
        <Text fontSize="sm" color="gray.500">
          已选 {selectedHosts.length} 台（最多 5 台）
        </Text>
      </HStack>
      <Box flex="1" w="100%">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart
            data={mergedData}
            margin={{ top: 10, right: 20, bottom: 20, left: 0 }}
          >
            <XAxis
              dataKey="ts"
              tickFormatter={formatTime}
              type="number"
              domain={["dataMin", "dataMax"]}
            />
            <YAxis unit=" ms" />
            <RTooltip
              labelFormatter={(l) => new Date(Number(l)).toLocaleTimeString()}
              formatter={(value: any) => [`${value} ms`, "latency"]}
            />
            <Legend />
            {selectedHosts.map((h, idx) => (
              <Line
                key={h}
                type="monotone"
                dataKey={h}
                stroke={COLORS[idx % COLORS.length]}
                dot={false}
                strokeWidth={2}
                isAnimationActive={false}
                connectNulls
              />
            ))}
          </LineChart>
        </ResponsiveContainer>
      </Box>
    </Box>
  );
}

export default TrendModal;
