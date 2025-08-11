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

  // X 轴 domain 规则：
  // - 主要目标：保持一个稳定且可预期的窗口，避免 start==end 造成 Recharts 计算异常
  // - 若数据跨度 >= 选定窗口：显示 [dataMax-window, dataMax]
  // - 若数据跨度 < 选定窗口：仍显示完整窗口 [dataMax-window, dataMax]（允许左侧留空）
  //   之前“去掉左侧空白”的策略在只有 1 个点或极少点时会让 domain 折叠为同一点，引发当前截图中的 X 轴异常
  // - 为了兼顾“数据基本填满窗口”且不浪费太多空白：当数据跨度覆盖窗口 60% 以上时再贴左端
  const xAxisDomain = useMemo(() => {
    const windowMs = timeRangeMinutes * 60 * 1000;
    if (!data || data.length === 0) {
      const end = Date.now();
      return [end - windowMs, end];
    }
    const tsList = data
      .map((d: any) => d.ts)
      .filter((v) => typeof v === "number" && v > 0);
    if (tsList.length === 0) {
      const end = Date.now();
      return [end - windowMs, end];
    }
    const dataMin = Math.min(...tsList);
    const dataMax = Math.max(...tsList);
    const span = dataMax - dataMin;
    const fullStart = dataMax - windowMs;
    // 如果跨度覆盖 >=60% 窗口并且最早数据点在 fullStart 之后，则可以贴到 dataMin 以减少空白
    if (span >= windowMs * 0.6 && dataMin > fullStart) {
      return [dataMin, dataMax];
    }
    // 否则使用标准窗口，保证不出现 start==end
    return [fullStart, dataMax];
  }, [timeRangeMinutes, data]);

  // 计算延迟 Y 轴 domain，避免所有值相等时 Recharts 产生异常刻度（之前出现 1428571429 / 8571428571 这类数字）
  const latencyDomain = useMemo(() => {
    if (!data || data.length === 0 || selectedHosts.length === 0) return [0, 1];
    let minV = Infinity;
    let maxV = -Infinity;
    for (const row of data as any[]) {
      for (const h of selectedHosts) {
        const avg = row[`${h}:avg`];
        const range = row[`${h}:range`];
        if (Array.isArray(range) && range.length === 2) {
          if (typeof range[0] === "number") minV = Math.min(minV, range[0]);
          if (typeof range[1] === "number") maxV = Math.max(maxV, range[1]);
        }
        if (typeof avg === "number") {
          minV = Math.min(minV, avg);
          maxV = Math.max(maxV, avg);
        }
      }
    }
    if (minV === Infinity || maxV === -Infinity) return [0, 1];
    // 处理所有值相同或跨度极小
    if (maxV - minV < 0.001) {
      const v = maxV;
      const pad = v === 0 ? 1 : Math.max(1, v * 0.2);
      return [Math.max(0, v - pad), v + pad];
    }
    return [
      Math.floor(Math.max(0, minV - (maxV - minV) * 0.1)),
      Math.ceil(maxV + (maxV - minV) * 0.1),
    ];
  }, [data, selectedHosts]);
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
              scale="time"
              interval="preserveStartEnd" // 让首尾刻度保留，避免被压缩
              minTickGap={32}
              tickFormatter={(v) =>
                new Date(v).toLocaleTimeString("zh-CN", {
                  hour12: false,
                  hour: "2-digit",
                  minute: "2-digit",
                  second: "2-digit",
                })
              }
              tickSize={8}
              axisLine={{ strokeWidth: 1 }}
              allowDataOverflow={false}
            />
            {/* 左侧Y轴：延迟 */}
            <YAxis
              yAxisId="left"
              unit=" ms"
              domain={latencyDomain as any}
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
                        strokeWidth={0.3}
                        r={2}
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
