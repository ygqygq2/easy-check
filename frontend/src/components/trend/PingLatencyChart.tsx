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
  return (
    <Box w="100%" h="100%">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart
          data={data}
          margin={{ top: 4, right: 16, bottom: 4, left: 0 }}
        >
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis
            dataKey="ts"
            type="number"
            domain={["dataMin", "dataMax"]}
            tickFormatter={(v) => new Date(v).toLocaleTimeString()}
          />
          <YAxis unit=" ms" domain={["dataMin - 2", "dataMax + 2"]} />
          <Tooltip
            labelFormatter={(l) => new Date(Number(l)).toLocaleTimeString()}
          />
          {/* min-max 带状阴影 + min/max 细线 + avg 粗线 */}
          {selectedHosts.map((h) => (
            <>
              {/* 带状阴影：用两层 Area 叠加模拟，避免 baseValue 类型限制 */}
              <Area
                key={`${h}-min-area`}
                type="monotone"
                dataKey={`${h}:min`}
                stroke={undefined}
                fill={colorMap[h]}
                fillOpacity={0.06}
                isAnimationActive={false}
                connectNulls
              />
              <Area
                key={`${h}-max-area`}
                type="monotone"
                dataKey={`${h}:max`}
                stroke={undefined}
                fill={colorMap[h]}
                fillOpacity={0.06}
                isAnimationActive={false}
                connectNulls
              />
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
              />
              <Line
                key={`${h}-avg`}
                type="monotone"
                dataKey={`${h}:avg`}
                stroke={colorMap[h]}
                dot={false}
                strokeWidth={2}
                isAnimationActive={false}
                connectNulls
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
              />
            </>
          ))}
        </LineChart>
      </ResponsiveContainer>
      {/* 自定义图例（在 X 轴下方） */}
      {selectedHosts.length > 0 && (
        <Box mt="1" px="2">
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
