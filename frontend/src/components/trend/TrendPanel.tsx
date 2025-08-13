import { Box, Button, HStack, Text } from "@chakra-ui/react";
import { useEffect, useMemo, useState } from "react";

import { HostSeriesMap } from "@/types/series";

import TimeRangePicker, {
  QuickRange,
  TimeRange as PickerTimeRange,
} from "../time/TimeRangePicker";
import PingLatencyChart from "./PingLatencyChart";
import { TIME_RANGES, TimeRange as LegacyTimeRange } from "./TimeRangeSelector";

const COLORS = ["#3182ce", "#38a169", "#d69e2e", "#e53e3e", "#805ad5"];

interface Props {
  selectedHosts: string[];
  seriesMap: HostSeriesMap; // host -> [{ ts, min, avg, max, loss }]
  onLoadHistory: (host: string, minutes: number) => Promise<void> | void; // 复用父级 hook 的加载函数
  stepSeconds?: number | null; // 来自服务端的实际分辨率（秒）
}

// 将 host->points 的结构合并为 Recharts 友好的一维数组
// 每个点包含 ts 和 `${host}:min|avg|max|loss` 等字段
function useMerged(seriesMap: HostSeriesMap, hosts: string[]) {
  return useMemo(() => {
    // 预构建：每个主机一张 ts->point 的索引表，避免 O(n^2) 的 Array.find
    const hostIndex: Record<
      string,
      Map<
        number,
        { ts: number; min?: number; avg?: number; max?: number; loss?: number }
      >
    > = {};
    const tsSet = new Set<number>();

    for (const h of hosts) {
      const points = seriesMap[h] || [];
      const map = new Map<
        number,
        { ts: number; min?: number; avg?: number; max?: number; loss?: number }
      >();
      for (const p of points) {
        map.set(p.ts, p);
        tsSet.add(p.ts);
      }
      hostIndex[h] = map;
    }

    const allTs = Array.from(tsSet).sort((a, b) => a - b);

    const rows = new Array(allTs.length);
    for (let i = 0; i < allTs.length; i++) {
      const ts = allTs[i];
      const row: Record<string, unknown> = { ts };
      for (const h of hosts) {
        const p = hostIndex[h]?.get(ts);
        if (!p) continue;
        if (typeof p.avg === "number") row[`${h}:avg`] = p.avg;
        if (typeof p.loss === "number") row[`${h}:loss`] = p.loss;
        if (typeof p.min === "number" && typeof p.max === "number") {
          row[`${h}:range`] = [p.min, p.max];
        }
      }
      rows[i] = row;
    }
    return rows;
  }, [seriesMap, hosts]);
}

export default function TrendPanel({
  selectedHosts,
  seriesMap,
  onLoadHistory,
  stepSeconds,
}: Props) {
  const [selectedTimeRange, setSelectedTimeRange] = useState<LegacyTimeRange>(
    TIME_RANGES[0]
  ); // 默认最近10分钟
  const [customRange, setCustomRange] = useState<{
    start: number;
    end: number;
  } | null>(null);

  // 当时间范围变化时，为所有选中的主机重新加载对应时间范围的历史数据（清除自定义范围）
  useEffect(() => {
    setCustomRange(null);
    const loadDataForTimeRange = async () => {
      const promises = selectedHosts.map((host) =>
        onLoadHistory(host, selectedTimeRange.minutes)
      );
      await Promise.all(promises);
    };

    if (selectedHosts.length > 0) {
      loadDataForTimeRange();
    }
  }, [selectedTimeRange, selectedHosts, onLoadHistory]);

  // 仅按时间窗口过滤旧点，不做前端再采样；采样将迁移到后端步长决策
  const merged = useMerged(
    useMemo(() => {
      const cutoff = customRange
        ? customRange.start
        : Date.now() - selectedTimeRange.minutes * 60 * 1000;
      const filtered: HostSeriesMap = {};
      Object.keys(seriesMap).forEach((host) => {
        const arr = seriesMap[host] || [];
        filtered[host] = customRange
          ? arr.filter(
              (p) => p.ts >= customRange.start && p.ts <= customRange.end
            )
          : arr.filter((p) => p.ts >= cutoff);
      });
      return filtered;
    }, [seriesMap, selectedTimeRange, customRange]),
    selectedHosts
  );

  const colorMap = useMemo(() => {
    const m: Record<string, string> = {};
    selectedHosts.forEach((h, i) => (m[h] = COLORS[i % COLORS.length]));
    return m;
  }, [selectedHosts]);
  const hasData = merged.length > 0;

  const humanStep = useMemo(() => {
    if (!stepSeconds || stepSeconds <= 0) return "自动";
    if (stepSeconds < 60) return `${stepSeconds}s`;
    if (stepSeconds % 3600 === 0) return `${stepSeconds / 3600}h`;
    if (stepSeconds % 60 === 0) return `${stepSeconds / 60}m`;
    return `${stepSeconds}s`;
  }, [stepSeconds]);

  // 将旧的 TIME_RANGES 映射为 TimeRangePicker 的快捷选项
  const quickRanges: QuickRange[] = useMemo(() => {
    const toQuick = (minutes: number): QuickRange => {
      if (minutes % 1440 === 0) {
        return {
          label: `最近${minutes / 1440}天`,
          amount: minutes / 1440,
          unit: "d",
        };
      }
      if (minutes % 60 === 0) {
        return {
          label: `最近${minutes / 60}小时`,
          amount: minutes / 60,
          unit: "h",
        };
      }
      return { label: `最近${minutes}分钟`, amount: minutes, unit: "m" };
    };
    return TIME_RANGES.map((r) => toQuick(r.minutes));
  }, []);

  // 计算 TimeRangePicker 的按钮文本与当前值
  const pickerButtonLabel = useMemo(() => {
    if (customRange) {
      return `${new Date(customRange.start).toLocaleString()} → ${new Date(
        customRange.end
      ).toLocaleString()}`;
    }
    return selectedTimeRange.label;
  }, [customRange, selectedTimeRange]);

  const pickerValue: PickerTimeRange = useMemo(() => {
    const now = Date.now();
    if (customRange) return { from: customRange.start, to: customRange.end };
    return { from: now - selectedTimeRange.minutes * 60_000, to: now };
  }, [customRange, selectedTimeRange]);

  return (
    <Box w="100%" h="100%" display="flex" flexDirection="column">
      <HStack gap="4" px="2" py="1" justify="space-between">
        <HStack gap="4">
          <Text>历史趋势</Text>
          <Text fontSize="xs" color="gray.500">
            分辨率: {humanStep}
          </Text>
          <Text fontSize="sm" color="gray.500">
            已选 {selectedHosts.length} 台（最多 5 台）
          </Text>
        </HStack>
        <HStack gap="2">
          {customRange ? (
            <Button
              size="xs"
              variant="outline"
              onClick={() => setCustomRange(null)}
            >
              重置
            </Button>
          ) : null}
          <TimeRangePicker
            value={pickerValue}
            quickRanges={quickRanges}
            buttonLabel={pickerButtonLabel}
            onApply={async (range) => {
              const now = Date.now();
              const nearNow = Math.abs(range.to - now) <= 60_000; // 允许 1 分钟误差
              const minutes = Math.max(
                1,
                Math.round((range.to - range.from) / 60_000)
              );

              if (nearNow) {
                // 视为相对时间：匹配预设时间段
                const match = TIME_RANGES.find((r) => r.minutes === minutes);
                if (match) {
                  setSelectedTimeRange(match);
                  return; // useEffect 会据此触发加载并清除自定义范围
                }
              }

              // 否则视为绝对时间
              setCustomRange({ start: range.from, end: range.to });
              const mins = Math.max(
                1,
                Math.round((range.to - range.from) / 60000)
              );
              const tasks = selectedHosts.map((h) => onLoadHistory(h, mins));
              await Promise.all(tasks);
            }}
          />
        </HStack>
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
            customRange={customRange}
            stepSeconds={stepSeconds}
            onZoom={async ({ start, end }) => {
              setCustomRange({ start, end });
              // 拉框放大时，按该范围重载数据（让服务端对齐网格与步长）
              const mins = Math.max(1, Math.round((end - start) / 60000));
              const tasks = selectedHosts.map((h) => onLoadHistory(h, mins));
              await Promise.all(tasks);
            }}
          />
        </Box>
      )}
    </Box>
  );
}
