"use client";
import {
  GetHosts,
  GetStatusWithHosts,
  GetHistoryWithHosts,
} from "@bindings/easy-check/internal/services/appservice";
import { Box, Grid, GridItem, Stack } from "@chakra-ui/react";
import { useEffect, useState } from "react";

import TrendPanel from "@/components/trend/TrendPanel";
import { toaster } from "@/components/ui/toaster";
import { Host, HostStatusMap } from "@/types/host";
import { HostSeriesMap, SeriesPoint } from "@/types/series";

import { PaginationControls } from "../components/PaginationControls";
import { HostList } from "./home/components/HostList";
import { RefreshIntervalSelector } from "./home/components/RefreshIntervalSelector";
import { SearchBar } from "./home/components/SearchBar";

export function Page() {
  const pageSize = 10;
  const [page, setPage] = useState(1);
  const [hosts, setHosts] = useState<Host[]>([]);
  const [total, setTotal] = useState(0);
  const [statusData, setStatusData] = useState<HostStatusMap>(new Map());
  const [refreshInterval, setRefreshInterval] = useState<number | null>(10000); // 默认10秒刷新
  const [searchTerm, setSearchTerm] = useState("");
  const [displayedHosts, setDisplayedHosts] = useState<Host[]>([]);
  const [selectedHosts, setSelectedHosts] = useState<string[]>([]);
  const [historyMap, setHistoryMap] = useState<HostSeriesMap>({});

  const fetchAndSetHosts = async (page: number, searchTerm: string) => {
    try {
      const res = (await GetHosts(page, pageSize, searchTerm)) || {
        hosts: [],
        total: 0,
      };
      const { hosts = [], total = 0 } = res;

      // 更新主机列表和分页总数
      setHosts(hosts);
      setTotal(total);

      // 初始化 latencyData 为 null
      const initialLatencyData: HostStatusMap = new Map();
      hosts.forEach((host) => {
        initialLatencyData.set(host.host, {
          description: host.description,
          latency: null,
          status: "RECOVERY",
          sent: false,
        });
      });
      setStatusData(initialLatencyData);

      // 更新显示的主机列表
      setDisplayedHosts(hosts);
    } catch (err) {
      console.error("Error fetching hosts:", err);
    }
  };

  useEffect(() => {
    fetchAndSetHosts(page, searchTerm);
  }, [page]);

  useEffect(() => {
    if (!searchTerm) {
      setDisplayedHosts(hosts || []);
    } else {
      // 在搜索时清空选中的主机，避免选中不可见主机导致的渲染循环
      setSelectedHosts([]);

      const lowercasedFilter = searchTerm.toLowerCase();
      const filtered = (hosts || []).filter(
        (host) =>
          host.host.toLowerCase().includes(lowercasedFilter) ||
          (host.description &&
            host.description.toLowerCase().includes(lowercasedFilter))
      );
      setDisplayedHosts(filtered);
    }
  }, [searchTerm, hosts]);

  // 获取历史数据的函数
  const fetchHistoryData = async (
    hostNames: string[],
    timeRangeMinutes: number
  ) => {
    try {
      const now = Date.now();
      const startTime = now - timeRangeMinutes * 60 * 1000;
      const endTime = now;

      // 根据时间范围计算步长（秒）
      let step = 60; // 默认1分钟
      if (timeRangeMinutes <= 60) step = 60; // 1小时内，1分钟间隔
      else if (timeRangeMinutes <= 360) step = 300; // 6小时内，5分钟间隔
      else if (timeRangeMinutes <= 1440) step = 900; // 1天内，15分钟间隔
      else if (timeRangeMinutes <= 10080) step = 3600; // 1周内，1小时间隔
      else step = 7200; // 更长时间，2小时间隔

      const historyRes = await GetHistoryWithHosts(
        hostNames,
        startTime,
        endTime,
        step
      );
      return historyRes;
    } catch (err) {
      console.error("Error fetching history data:", err);
      return null;
    }
  };

  useEffect(() => {
    const fetchHostsStatus = async () => {
      if (hosts?.length === 0) {
        setStatusData(new Map());
        return;
      }
      try {
        const hostNames = hosts.map((host) => host.host);
        if (hostNames?.length === 0) {
          setStatusData(new Map());
          return;
        }
        const res = await GetStatusWithHosts(hostNames);
        const statusHosts = res?.hosts || [];
        const statusList: HostStatusMap = new Map();
        const now = Date.now();
        const windowMs = 30 * 24 * 60 * 60 * 1000; // 30天窗口，支持最长时间范围
        let nextHistory: HostSeriesMap = { ...historyMap };

        statusHosts.forEach((statusHost) => {
          statusList.set(statusHost.host, {
            description: statusHost.host,
            latency: statusHost.avg_latency || null,
            status: statusHost.status === "ALERT" ? "ALERT" : "RECOVERY",
            sent: false,
          });

          // 追加当前最新的数据点到已有的历史数据中
          const point: SeriesPoint = {
            ts: now,
            min:
              typeof statusHost.min_latency === "number"
                ? statusHost.min_latency
                : undefined,
            avg:
              typeof statusHost.avg_latency === "number"
                ? statusHost.avg_latency
                : undefined,
            max:
              typeof statusHost.max_latency === "number"
                ? statusHost.max_latency
                : undefined,
            loss:
              typeof statusHost.packet_loss === "number"
                ? statusHost.packet_loss
                : undefined,
          };

          const hostName = statusHost.host;
          const existingData = nextHistory[hostName] || [];

          if (
            existingData.length === 0 ||
            existingData[existingData.length - 1].ts !== now
          ) {
            existingData.push(point);
          } else {
            existingData[existingData.length - 1] = point;
          }

          const cutoff = now - windowMs;
          nextHistory[statusHost.host] = existingData
            .filter((p) => p.ts >= cutoff)
            .slice(-10000); // 支持长时间数据
        });

        setStatusData(statusList);
        setHistoryMap(nextHistory);
      } catch (err) {
        console.error("Error fetching latency data:", err);
      }
    };

    if (hosts.length > 0) {
      fetchHostsStatus();
    } else {
      setStatusData(new Map());
    }

    let intervalId: NodeJS.Timeout | null = null;
    if (refreshInterval && hosts.length > 0) {
      intervalId = setInterval(fetchHostsStatus, refreshInterval);
    }
    return () => {
      if (intervalId) clearInterval(intervalId);
    };
  }, [hosts, refreshInterval]);

  // 选择主机（最多5个），自动打开趋势图
  const toggleHost = (host: string) => {
    // 额外防护：若该主机处于 ALERT 或无数据，不允许选中
    const s = statusData.get(host);
    const disabled =
      !s || s.status === "ALERT" || typeof s.latency !== "number";
    if (disabled) {
      toaster.create({
        title: "该主机当前不可选",
        description: s?.status === "ALERT" ? "主机告警中" : "暂无数据",
        type: "warning",
      });
      return;
    }
    setSelectedHosts((prev) => {
      const exists = prev.includes(host);
      let next = exists ? prev.filter((h) => h !== host) : [...prev, host];
      if (next.length > 5) {
        toaster.create({
          title: "最多选择5个主机",
          description: "已达到选择上限",
          type: "warning",
        });
        next = next.slice(0, 5);
      }
      // 自动打开趋势图（至少选择1个时）
      return next;
    });
  };

  // 当取消选择为0时，关闭弹窗
  // 无需弹窗联动

  // 当选中的主机变化时，加载历史数据
  useEffect(() => {
    const loadHistoryForSelectedHosts = async () => {
      if (selectedHosts.length === 0) {
        return; // 没有选中主机，不需要加载历史数据
      }

      try {
        // 只为新选中的主机（没有历史数据的）获取历史数据
        const hostsNeedingHistory = selectedHosts.filter((hostName) => {
          const existingData = historyMap[hostName] || [];
          // 如果该主机的历史数据点少于5个，说明需要加载历史数据
          return existingData.length < 5;
        });

        if (hostsNeedingHistory.length === 0) {
          return; // 所有选中的主机都已经有足够的历史数据了
        }

        console.log("Loading history data for hosts:", hostsNeedingHistory);

        // 获取30分钟的历史数据
        const historyRes = await fetchHistoryData(hostsNeedingHistory, 30);

        if (historyRes?.hosts) {
          setHistoryMap((prevHistoryMap) => {
            const nextHistory: HostSeriesMap = { ...prevHistoryMap };

            historyRes.hosts.forEach((hostData) => {
              const hostName = hostData.host;
              const historicalPoints: SeriesPoint[] = [];

              // 转换历史数据格式
              if (hostData.series) {
                // 假设指标名为 avg_latency, min_latency, max_latency, packet_loss
                const avgData = hostData.series["avg_latency"] || [];
                const minData = hostData.series["min_latency"] || [];
                const maxData = hostData.series["max_latency"] || [];
                const lossData = hostData.series["packet_loss"] || [];

                // 创建时间戳到数据点的映射
                const pointsMap: { [ts: number]: SeriesPoint } = {};

                avgData.forEach((point) => {
                  if (!pointsMap[point.timestamp]) {
                    pointsMap[point.timestamp] = { ts: point.timestamp };
                  }
                  pointsMap[point.timestamp].avg = point.value;
                });

                minData.forEach((point) => {
                  if (!pointsMap[point.timestamp]) {
                    pointsMap[point.timestamp] = { ts: point.timestamp };
                  }
                  pointsMap[point.timestamp].min = point.value;
                });

                maxData.forEach((point) => {
                  if (!pointsMap[point.timestamp]) {
                    pointsMap[point.timestamp] = { ts: point.timestamp };
                  }
                  pointsMap[point.timestamp].max = point.value;
                });

                lossData.forEach((point) => {
                  if (!pointsMap[point.timestamp]) {
                    pointsMap[point.timestamp] = { ts: point.timestamp };
                  }
                  pointsMap[point.timestamp].loss = point.value;
                });

                // 转换为数组并排序
                historicalPoints.push(
                  ...Object.values(pointsMap).sort((a, b) => a.ts - b.ts)
                );
              }

              // 只在该主机没有足够历史数据时才合并
              const existingData = nextHistory[hostName] || [];
              if (existingData.length < 5) {
                const mergedData = [...historicalPoints];

                // 添加现有的实时数据点（过滤重复时间戳）
                existingData.forEach((point) => {
                  if (
                    !mergedData.find((p) => Math.abs(p.ts - point.ts) < 30000)
                  ) {
                    // 30秒内认为是重复
                    mergedData.push(point);
                  }
                });

                // 排序并去重
                mergedData.sort((a, b) => a.ts - b.ts);

                const windowMs = 30 * 24 * 60 * 60 * 1000;
                const cutoff = Date.now() - windowMs;
                nextHistory[hostName] = mergedData
                  .filter((p) => p.ts >= cutoff)
                  .slice(-10000);
              }
            });

            return nextHistory;
          });

          console.log("History data loaded for hosts:", hostsNeedingHistory);
        }
      } catch (err) {
        console.error("Error loading history data for selected hosts:", err);
      }
    };

    loadHistoryForSelectedHosts();
  }, [selectedHosts.join(",")]); // 使用字符串形式避免数组引用变化导致的重复执行

  const handleBackendSearch = async (searchTerm: string) => {
    setPage(1); // 重置到第一页
    await fetchAndSetHosts(1, searchTerm);
  };

  return (
    <Box h="100vh" display="flex" flexDirection="column" p="2" gap="2">
      <Box flexShrink={0}>
        <Stack
          direction="row"
          gap="3"
          align="center"
          justify="space-between"
          mb="2"
          px="2"
        >
          <SearchBar
            searchTerm={searchTerm}
            setSearchTerm={setSearchTerm}
            onSearch={() => handleBackendSearch(searchTerm)}
          />
          <RefreshIntervalSelector
            refreshInterval={refreshInterval}
            onChange={setRefreshInterval}
          />
        </Stack>
        <HostList
          hosts={displayedHosts}
          statusData={statusData}
          selectedHosts={selectedHosts}
          onToggleHost={toggleHost}
        />
        <PaginationControls
          page={page}
          total={total}
          pageSize={pageSize}
          onPageChange={setPage}
        />
      </Box>
      <Box flex="1" minH={0} maxH="300px">
        <TrendPanel selectedHosts={selectedHosts} seriesMap={historyMap} />
      </Box>
    </Box>
  );
}
