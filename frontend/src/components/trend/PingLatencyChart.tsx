import { memo, useMemo } from "react";
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
  TooltipProps,
} from "recharts";
import { SeriesPoint } from "@/types/series";
import { useColorModeValue } from "@/components/ui/color-mode";

// 简化的 Tooltip 组件
const SimpleTooltip = ({ active, payload, label }: any) => {
  const tooltipBg = useColorModeValue(
    "rgba(255,255,255,0.95)",
    "rgba(45,55,72,0.95)"
  );
  const tooltipBorder = useColorModeValue("#e2e8f0", "#4a5568");

  if (!active || !payload || !payload.length) return null;

  // 只显示平均延迟和丢包率
  const items = payload
    .filter((item: any) => item.strokeWidth > 0)
    .filter((item: any) => {
      const dataKey = item.dataKey as string;
      return dataKey.includes(":avg") || dataKey.includes(":loss");
    });

  return (
    <div
      style={{
        fontSize: "12px",
        backgroundColor: tooltipBg,
        border: `1px solid ${tooltipBorder}`,
        borderRadius: "4px",
        boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
        padding: "8px",
      }}
    >
      <p style={{ margin: "0 0 4px 0" }}>
        {new Date(Number(label)).toLocaleString("zh-CN", { hour12: false })}
      </p>
      {items.map((item: any, index: number) => {
        const dataKey = item.dataKey as string;
        const numValue = Number(item.value);
        const hostName = dataKey.split(":")[0];
        const isLoss = dataKey.includes(":loss");
        const displayName = `${hostName} ${isLoss ? "丢包率" : "平均延迟"}`;
        const displayValue = `${numValue.toFixed(2)}${isLoss ? "%" : "ms"}`;

        return (
          <p
            key={index}
            style={{
              margin: "2px 0",
              color: item.color,
              display: "flex",
              justifyContent: "space-between",
            }}
          >
            <span>{displayName}:</span>
            <span style={{ marginLeft: "8px", fontWeight: "bold" }}>
              {displayValue}
            </span>
          </p>
        );
      })}
    </div>
  );
};

interface Props {
  data: SeriesPoint[];
  selectedHosts: string[];
  // host => color
  colorMap: Record<string, string>;
  // 时间范围（分钟）
  timeRangeMinutes: number;
}

const PingLatencyChart = memo(function PingLatencyChart({
  data,
  selectedHosts,
  colorMap,
  timeRangeMinutes,
}: Props) {
  const showDots = (data?.length || 0) < 3;

  // 主题相关的颜色
  const axisColor = useColorModeValue("gray.600", "gray.400");
  const tooltipBg = useColorModeValue(
    "rgba(255,255,255,0.95)",
    "rgba(45,55,72,0.95)"
  );
  const tooltipBorder = useColorModeValue("#e2e8f0", "#4a5568");

  // 根据时间范围计算X轴domain
  const xAxisDomain = useMemo(() => {
    const now = Date.now();
    const start = now - timeRangeMinutes * 60 * 1000;
    return [start, now];
  }, [timeRangeMinutes]);
  return (
    <Box w="100%" h="100%" display="flex" flexDirection="column">
      <style>
        {`
          /* Recharts轴标签字体大小和颜色 - 使用CSS确保在所有环境下生效 */
          .recharts-cartesian-axis-tick-value {
            font-size: 11px !important;
            fill: ${axisColor} !important;
          }
          .recharts-cartesian-axis-tick text {
            font-size: 11px !important;
            fill: ${axisColor} !important;
          }
        `}
      </style>
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
              domain={xAxisDomain}
              tickFormatter={(v) =>
                new Date(v).toLocaleTimeString("zh-CN", { hour12: false })
              }
              tickSize={8}
              axisLine={{ strokeWidth: 1 }}
            />
            {/* 左侧Y轴：延迟 */}
            <YAxis
              yAxisId="left"
              unit=" ms"
              domain={["dataMin", "dataMax"]}
              allowDecimals={false}
              width={50}
              tickSize={8}
              axisLine={{ strokeWidth: 1 }}
            />
            {/* 右侧Y轴：丢包率 */}
            <YAxis
              yAxisId="right"
              orientation="right"
              unit="%"
              domain={[0, "dataMax"]}
              allowDecimals={false}
              width={50}
              tickSize={8}
              axisLine={{ strokeWidth: 1 }}
            />
            <Tooltip content={<SimpleTooltip />} />
            {/* min-max 带状阴影 + min/max 细线 + avg 粗线 */}
            {selectedHosts.map((h) => (
              <>
                {/* 延迟数据：带状阴影 - 隐藏在tooltip中的显示 */}
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
