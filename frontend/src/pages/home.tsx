"use client";
import {
  GetHosts,
  GetStatusWithHosts,
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
        const windowMs = 10 * 60 * 1000; // 10分钟窗口
        const nextHistory: HostSeriesMap = { ...historyMap };
        statusHosts.forEach((statusHost) => {
          statusList.set(statusHost.host, {
            description: statusHost.host,
            latency: statusHost.avg_latency || null,
            status: statusHost.status === "ALERT" ? "ALERT" : "RECOVERY",
            sent: false,
          });

          // 追加历史点（min/avg/max/loss）
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
          const arr = nextHistory[statusHost.host]
            ? [...nextHistory[statusHost.host]]
            : [];
          if (arr.length === 0 || arr[arr.length - 1].ts !== now) {
            arr.push(point);
          } else {
            arr[arr.length - 1] = point;
          }
          const cutoff = now - windowMs;
          nextHistory[statusHost.host] = arr
            .filter((p) => p.ts >= cutoff)
            .slice(-400);
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
