import { memo } from "react";
import { Box, HStack, Text, Wrap, WrapItem } from "@chakra-ui/react";
import {
  LineChart,
  Area,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { SeriesPoint } from "@/types/series";

interface Props {
  data: SeriesPoint[];
  selectedHosts: string[];
  // host => color
  colorMap: Record<string, string>;
}

const PingLatencyChart = memo(function PingLatencyChart({
  data,
  selectedHosts,
  colorMap,
}: Props) {
  const showDots = (data?.length || 0) < 3;
  return (
    <Box w="100%" h="100%" display="flex" flexDirection="column">
      <Box flex="1" minH={0}>
        <ResponsiveContainer width="100%" height="100%">
          <LineChart
            data={data}
            margin={{ top: 10, right: 30, bottom: 10, left: 30 }}
          >
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis
              dataKey="ts"
              type="number"
              domain={[
                (dataMin: number) => (isFinite(dataMin) ? dataMin - 1000 : 0),
                (dataMax: number) =>
                  isFinite(dataMax) ? dataMax + 1000 : 1000,
              ]}
              tickFormatter={(v) => new Date(v).toLocaleTimeString()}
            />
            {/* 左侧Y轴：延迟 */}
            <YAxis
              yAxisId="left"
              unit=" ms"
              domain={["dataMin", "dataMax"]}
              allowDecimals={false}
              width={50}
            />
            {/* 右侧Y轴：丢包率 */}
            <YAxis
              yAxisId="right"
              orientation="right"
              unit="%"
              domain={[0, "dataMax"]}
              allowDecimals={false}
              width={50}
            />
            <Tooltip
              labelFormatter={(l) => new Date(Number(l)).toLocaleTimeString()}
            />
            {/* min-max 带状阴影 + min/max 细线 + avg 粗线 */}
            {selectedHosts.map((h) => (
              <>
                {/* 延迟数据：带状阴影 */}
                <Area
                  key={`${h}-min-area`}
                  type="monotone"
                  dataKey={`${h}:min`}
                  stroke={undefined}
                  fill={colorMap[h]}
                  fillOpacity={0.15}
                  isAnimationActive={false}
                  connectNulls
                  yAxisId="left"
                />
                <Area
                  key={`${h}-max-area`}
                  type="monotone"
                  dataKey={`${h}:max`}
                  stroke={undefined}
                  fill={colorMap[h]}
                  fillOpacity={0.15}
                  isAnimationActive={false}
                  connectNulls
                  yAxisId="left"
                />
                {/* 延迟数据：线条 */}
                <Line
                  key={`${h}-min`}
                  type="monotone"
                  dataKey={`${h}:min`}
                  stroke={colorMap[h]}
                  strokeDasharray="4 4"
                  dot={false}
                  strokeWidth={1}
                  isAnimationActive={false}
                  connectNulls
                  yAxisId="left"
                />
                <Line
                  key={`${h}-avg`}
                  type="monotone"
                  dataKey={`${h}:avg`}
                  stroke={colorMap[h]}
                  dot={showDots ? { r: 2 } : false}
                  strokeWidth={2}
                  isAnimationActive={false}
                  connectNulls
                  yAxisId="left"
                />
                <Line
                  key={`${h}-max`}
                  type="monotone"
                  dataKey={`${h}:max`}
                  stroke={colorMap[h]}
                  strokeDasharray="4 4"
                  dot={false}
                  strokeWidth={1}
                  isAnimationActive={false}
                  connectNulls
                  yAxisId="left"
                />
                {/* 丢包率数据：线条 */}
                <Line
                  key={`${h}-loss`}
                  type="monotone"
                  dataKey={`${h}:loss`}
                  stroke={colorMap[h]}
                  strokeDasharray="2 2"
                  dot={showDots ? { r: 1 } : false}
                  strokeWidth={1.5}
                  isAnimationActive={false}
                  connectNulls
                  yAxisId="right"
                />
              </>
            ))}
          </LineChart>
        </ResponsiveContainer>
      </Box>
      {/* 自定义图例（在 X 轴下方，固定高度） */}
      {selectedHosts.length > 0 && (
        <Box mt="1" px="2" flexShrink={0}>
          <Wrap gap="12px">
            {selectedHosts.map((h) => (
              <WrapItem key={`legend-${h}`}>
                <HStack gap="1.5">
                  <Box
                    w="10px"
                    h="10px"
                    borderRadius="full"
                    bg={colorMap[h]}
                    boxShadow="inset 0 0 0 1px rgba(0,0,0,0.25)"
                  />
                  <Text fontSize="xs" color="gray.600">
                    {h}
                  </Text>
                </HStack>
              </WrapItem>
            ))}
          </Wrap>
        </Box>
      )}
    </Box>
  );
});

export default PingLatencyChart;
