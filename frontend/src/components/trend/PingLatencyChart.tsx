import { Box } from "@chakra-ui/react";
import { memo, useCallback, useMemo, useRef, useState } from "react";
import {
  Area,
  CartesianGrid,
  ComposedChart,
  Line,
  ReferenceArea,
  ResponsiveContainer,
  Scatter,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import { useColorModeValue } from "@/components/ui/color-mode";
import { useTooltipOptimization } from "@/hooks/useTooltipOptimization";
import { SeriesPoint } from "@/types/series";

import ChartLegend from "./ChartLegend";
import SimpleTooltip from "./SimpleTooltip";
import { getPacketLossColor } from "./smokeping-colors";

interface Props {
  data: SeriesPoint[];
  selectedHosts: string[];
  // host => color
  colorMap: Record<string, string>;
  // 时间范围（分钟）
  timeRangeMinutes: number;
  // 可选：自定义时间范围（毫秒），用于放大视图
  customRange?: { start: number; end: number } | null;
  // 当前分辨率（秒），用于生成更合理的刻度
  stepSeconds?: number | null;
  // 放大回调（使用鼠标拖动触发）
  onZoom?: (range: { start: number; end: number }) => void;
}

const PingLatencyChart = memo(function PingLatencyChart({
  data,
  selectedHosts,
  colorMap,
  timeRangeMinutes,
  customRange,
  stepSeconds,
  onZoom,
}: Props) {
  const [dragStart, setDragStart] = useState<number | null>(null);
  const [dragEnd, setDragEnd] = useState<number | null>(null);

  const chartRef = useRef<HTMLDivElement>(null);
  const _showDots = (data?.length || 0) < 3;

  // 主题相关的颜色
  const axisColor = useColorModeValue("gray.600", "gray.400");
  const selectionColor = useColorModeValue("#2B6CB0", "#63B3ED");

  // X 轴 domain 规则：
  // - 主要目标：保持一个稳定且可预期的窗口，避免 start==end 造成 Recharts 计算异常
  // - 若数据跨度 >= 选定窗口：显示 [dataMax-window, dataMax]
  // - 若数据跨度 < 选定窗口：仍显示完整窗口 [dataMax-window, dataMax]（允许左侧留空）
  //   之前“去掉左侧空白”的策略在只有 1 个点或极少点时会让 domain 折叠为同一点，引发当前截图中的 X 轴异常
  // - 为了兼顾“数据基本填满窗口”且不浪费太多空白：当数据跨度覆盖窗口 60% 以上时再贴左端
  const xAxisDomain = useMemo(() => {
    if (customRange && customRange.start && customRange.end) {
      return [customRange.start, customRange.end] as [number, number];
    }
    const windowMs = timeRangeMinutes * 60 * 1000;
    if (!data || data.length === 0) {
      const end = Date.now();
      return [end - windowMs, end];
    }
    const tsList = data
      .map((d: SeriesPoint) => d.ts)
      .filter((v) => typeof v === "number" && v > 0);
    if (tsList.length === 0) {
      const end = Date.now();
      return [end - windowMs, end];
    }
    const dataMin = Math.min(...tsList);
    const dataMax = Math.max(...tsList);
    const span = dataMax - dataMin;
    const fullStart = dataMax - windowMs;
    if (span >= windowMs * 0.6 && dataMin > fullStart) {
      return [dataMin, dataMax];
    }
    return [fullStart, dataMax];
  }, [timeRangeMinutes, data, customRange]);

  // 使用优化的 tooltip hook
  const {
    tooltipState,
    handleMouseMove: optimizedMouseMove,
    handleMouseLeave: optimizedMouseLeave,
  } = useTooltipOptimization({
    data: data || [],
    selectedHosts,
    xAxisDomain: xAxisDomain as [number, number],
    debounceMs: 16, // ~60fps
  });

  // 基于步长生成更友好的刻度
  const xTicks = useMemo(() => {
    const [start, end] = xAxisDomain as [number, number];
    const spanSec = Math.max(1, Math.floor((end - start) / 1000));
    const desired = 8; // 期望 8 个刻度
    let base = Math.max(1, Math.floor(spanSec / desired));
    if (stepSeconds && stepSeconds > 0) {
      const k = Math.max(1, Math.round(base / stepSeconds));
      base = k * stepSeconds;
    }
    const ticks: number[] = [];
    const startSec = Math.floor(start / 1000 / base) * base;
    const endSec = Math.ceil(end / 1000 / base) * base;
    for (let t = startSec; t <= endSec; t += base) {
      ticks.push(t * 1000);
      if (ticks.length > 60) break;
    }
    return ticks;
  }, [xAxisDomain, stepSeconds]);

  // 优化的鼠标移动处理
  const _handleChartMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!chartRef.current) return;

      const chartRect = chartRef.current.getBoundingClientRect();
      const mouseX = e.clientX - chartRect.left;
      const mouseY = e.clientY - chartRect.top;

      // 考虑图表的 margin
      const margin = { top: 10, right: 30, bottom: 10, left: 30 };
      const chartWidth = chartRect.width - margin.left - margin.right;
      const adjustedMouseX = mouseX - margin.left;

      // 确保鼠标在有效区域内
      if (
        adjustedMouseX >= 0 &&
        adjustedMouseX <= chartWidth &&
        mouseY >= margin.top &&
        mouseY <= chartRect.height - margin.bottom
      ) {
        optimizedMouseMove(adjustedMouseX, mouseY, chartWidth);
      } else {
        optimizedMouseLeave();
      }
    },
    [optimizedMouseMove, optimizedMouseLeave]
  );

  // 使用 Recharts 最新事件结构：优先从 activePayload 获取 ts
  const onChartMouseDown = useCallback(
    (e: {
      activePayload?: Array<{ payload?: SeriesPoint }>;
      activeLabel?: string | number;
    }) => {
      const ts = e?.activePayload?.[0]?.payload?.ts;
      if (typeof ts === "number") {
        setDragStart(ts);
        setDragEnd(null);
        return;
      }
      if (typeof e?.activeLabel === "number") {
        setDragStart(e.activeLabel);
        setDragEnd(null);
      } else if (typeof e?.activeLabel === "string") {
        const parsed = parseFloat(e.activeLabel);
        if (!isNaN(parsed)) {
          setDragStart(parsed);
          setDragEnd(null);
        }
      }
    },
    []
  );

  // 鼠标移动：更新拖拽终点 + 驱动 tooltip
  const onChartMouseMove = useCallback(
    (e: {
      activePayload?: Array<{ payload?: SeriesPoint }>;
      activeLabel?: string | number;
      activeCoordinate?: { x: number; y: number };
    }) => {
      if (dragStart !== null) {
        const ts = e?.activePayload?.[0]?.payload?.ts;
        if (typeof ts === "number") {
          setDragEnd(ts);
        } else if (typeof e?.activeLabel === "number") {
          setDragEnd(e.activeLabel);
        } else if (typeof e?.activeLabel === "string") {
          const parsed = parseFloat(e.activeLabel);
          if (!isNaN(parsed)) {
            setDragEnd(parsed);
          }
        }
      }

      // 驱动优化后的 tooltip
      const ax = e?.activeCoordinate?.x;
      const ay = e?.activeCoordinate?.y;
      if (typeof ax === "number" && typeof ay === "number") {
        const rect = chartRef.current?.getBoundingClientRect();
        const chartWidth = rect ? rect.width - 30 - 30 : 0; // 与 margin 匹配
        optimizedMouseMove(ax, ay, chartWidth);
      }
    },
    [dragStart, optimizedMouseMove]
  );

  const onChartMouseUp = useCallback(() => {
    if (dragStart && dragEnd && onZoom) {
      const start = Math.min(dragStart, dragEnd);
      const end = Math.max(dragStart, dragEnd);
      if (end - start > 5 * 1000) {
        onZoom({ start, end });
      }
    }
    setDragStart(null);
    setDragEnd(null);
  }, [dragStart, dragEnd, onZoom]);

  // 计算延迟 Y 轴 domain，避免所有值相等时 Recharts 产生异常刻度（之前出现 1428571429 / 8571428571 这类数字）
  const latencyDomain = useMemo(() => {
    if (!data || data.length === 0 || selectedHosts.length === 0) return [0, 1];
    let minV = Infinity;
    let maxV = -Infinity;
    for (const row of data) {
      for (const h of selectedHosts) {
        const avg = (row as unknown as Record<string, unknown>)[`${h}:avg`];
        const range = (row as unknown as Record<string, unknown>)[`${h}:range`];
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
      <Box flex="1" minH={0} position="relative">
        <div
          ref={chartRef}
          style={{ width: "100%", height: "100%", position: "relative" }}
        >
          <ResponsiveContainer width="100%" height="100%">
            <ComposedChart
              data={data}
              margin={{ top: 10, right: 30, bottom: 10, left: 30 }}
              onMouseDown={onChartMouseDown}
              onMouseMove={onChartMouseMove}
              onMouseUp={onChartMouseUp}
              onMouseLeave={optimizedMouseLeave}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis
                dataKey="ts"
                type="number"
                domain={xAxisDomain}
                scale="time"
                ticks={xTicks}
                interval={0}
                minTickGap={32}
                tickFormatter={(v) => {
                  const d = new Date(v);
                  const [start, end] = xAxisDomain as [number, number];
                  const crossDays =
                    new Date(start).toDateString() !==
                    new Date(end).toDateString();
                  if (crossDays) {
                    return d.toLocaleString("zh-CN", {
                      month: "2-digit",
                      day: "2-digit",
                      hour12: false,
                      hour: "2-digit",
                      minute: "2-digit",
                    });
                  }
                  return d.toLocaleTimeString("zh-CN", {
                    hour12: false,
                    hour: "2-digit",
                    minute: "2-digit",
                    second: "2-digit",
                  });
                }}
                tickSize={8}
                axisLine={{ strokeWidth: 1 }}
                allowDataOverflow={false}
              />
              {/* 拖拽选择区域可视化 - 主题色阴影 */}
              {dragStart !== null && dragEnd !== null && (
                <ReferenceArea
                  x1={Math.min(dragStart!, dragEnd!)}
                  x2={Math.max(dragStart!, dragEnd!)}
                  stroke={selectionColor}
                  strokeOpacity={0.85}
                  strokeWidth={1.2}
                  fill={selectionColor}
                  fillOpacity={0.12}
                  strokeDasharray="4 2"
                />
              )}
              {/* 左侧Y轴：延迟 */}
              <YAxis
                yAxisId="left"
                unit=" ms"
                domain={latencyDomain as [number, number]}
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
              {/* 禁用默认 Tooltip，我们使用自定义的 */}
              <Tooltip content={() => null} />
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
              {/* 丢包率标记：细小水平线段（更接近 smokeping 的视觉） */}
              {selectedHosts.map((host) => (
                <Scatter
                  key={`${host}-loss-markers`}
                  dataKey="avg"
                  data={data
                    .map((p: SeriesPoint) => ({
                      ts: p.ts,
                      avg: (p as unknown as Record<string, unknown>)[
                        `${host}:avg`
                      ],
                      loss: (p as unknown as Record<string, unknown>)[
                        `${host}:loss`
                      ],
                    }))
                    .filter(
                      (p) =>
                        typeof p.avg === "number" &&
                        typeof p.loss === "number" &&
                        p.loss > 0
                    )}
                  yAxisId="left"
                  // 自定义形状：在 (cx, cy) 位置画一小段水平线
                  shape={(props: {
                    cx?: number;
                    cy?: number;
                    payload?: { loss?: number };
                  }) => {
                    const { cx, cy, payload } = props;
                    if (typeof cx !== "number" || typeof cy !== "number")
                      return <g />;
                    const loss = payload?.loss ?? 0;
                    const color = getPacketLossColor(loss, colorMap[host]);
                    const half = 2; // 线段半长，整体约 4px
                    return (
                      <g>
                        <line
                          x1={cx - half}
                          y1={cy}
                          x2={cx + half}
                          y2={cy}
                          stroke={color}
                          strokeWidth={3}
                          strokeLinecap="round"
                        />
                      </g>
                    );
                  }}
                />
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
        </div>

        {/* 自定义 Tooltip */}
        {tooltipState?.visible && (
          <div
            style={{
              position: "absolute",
              left: tooltipState.x + 10,
              top: tooltipState.y - 10,
              pointerEvents: "none",
              zIndex: 1000,
            }}
          >
            <SimpleTooltip
              active={true}
              label={tooltipState.timestamp}
              data={data}
              selectedHosts={selectedHosts}
              payload={[]}
            />
          </div>
        )}
      </Box>
      {/* 自定义图例（在 X 轴下方，固定高度） */}
      <ChartLegend selectedHosts={selectedHosts} colorMap={colorMap} />
    </Box>
  );
});

export default PingLatencyChart;
