import { memo, useMemo } from "react";
import { Box } from "@chakra-ui/react";
import {
  ComposedChart,
  Area,
  Line,
  Scatter,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from "recharts";
import { SeriesPoint } from "@/types/series";
import { useColorModeValue } from "@/components/ui/color-mode";
import { getPacketLossColor } from "./smokeping-colors";
import SimpleTooltip from "./SimpleTooltip";
import ChartLegend from "./ChartLegend";

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
          <ComposedChart
            data={data}
            margin={{ top: 10, right: 30, bottom: 10, left: 30 }}
          >
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis
              dataKey="ts"
              type="number"
              domain={xAxisDomain}
              tickCount={15}
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
            {/* 延迟范围阴影区域和平均线 */}
            {selectedHosts.map((h) => (
              <>
                {/* 延迟范围阴影 */}
                <Area
                  key={`${h}-range`}
                  type="monotone"
                  dataKey={`${h}:range`}
                  stroke="none"
                  fill={colorMap[h]}
                  fillOpacity={0.2}
                  isAnimationActive={false}
                  connectNulls
                  dot={false}
                  activeDot={false}
                  yAxisId="left"
                />
                {/* 平均延迟线 */}
                <Line
                  key={`${h}-avg`}
                  type="monotone"
                  dataKey={`${h}:avg`}
                  stroke={colorMap[h]}
                  dot={false}
                  strokeWidth={2}
                  isAnimationActive={false}
                  connectNulls
                  yAxisId="left"
                />
              </>
            ))}
            {/* 丢包率散点图 - 只在有丢包时显示 */}
            {selectedHosts.map((host) => (
              <Scatter
                key={`${host}-scatter`}
                dataKey="avg"
                data={data
                  .map((point: any) => ({
                    ts: point.ts,
                    avg: point[`${host}:avg`],
                    loss: point[`${host}:loss`],
                  }))
                  .filter(
                    (point) =>
                      point.avg !== undefined &&
                      point.loss !== undefined &&
                      point.loss > 0 // 只显示有丢包的点
                  )}
                fill="#8884d8"
                shape="square"
                yAxisId="left"
              >
                {data
                  .filter((point: any) => {
                    const loss = point[`${host}:loss`];
                    const avg = point[`${host}:avg`];
                    return avg !== undefined && loss !== undefined && loss > 0;
                  })
                  .map((point: any, index: number) => {
                    const loss = point[`${host}:loss`];
                    const color = getPacketLossColor(loss, colorMap[host]);
                    return (
                      <Cell
                        key={`scatter-${host}-${index}`}
                        fill={color}
                        stroke={color}
                        strokeWidth={0.5}
                        r={3} // 减小方框尺寸
                      />
                    );
                  })}
              </Scatter>
            ))}
            {/* 添加隐藏的丢包率线条用于工具提示 */}
            {selectedHosts.map((host) => (
              <Line
                key={`${host}-loss-tooltip`}
                type="monotone"
                dataKey={`${host}:loss`}
                stroke="transparent"
                strokeWidth={0}
                dot={false}
                isAnimationActive={false}
                connectNulls={false}
                yAxisId="right"
              />
            ))}
          </ComposedChart>
        </ResponsiveContainer>
      </Box>
      {/* 自定义图例（在 X 轴下方，固定高度） */}
      <ChartLegend selectedHosts={selectedHosts} colorMap={colorMap} />
    </Box>
  );
});

export default PingLatencyChart;
