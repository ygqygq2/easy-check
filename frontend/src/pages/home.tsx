"use client";
import {
  GetHosts,
  GetLatencyWithHosts,
} from "@bindings/easy-check/internal/services/appservice";
import { Box, Stack } from "@chakra-ui/react";
import { useEffect, useState } from "react";

import { Host } from "@/types/host";

import { PaginationControls } from "../components/PaginationControls";
import { HostList } from "./home/components/HostList";
import { RefreshIntervalSelector } from "./home/components/RefreshIntervalSelector";
import { SearchBar } from "./home/components/SearchBar";

export function Page() {
  const pageSize = 20;
  const [page, setPage] = useState(1);
  const [hosts, setHosts] = useState<Host[]>([]);
  const [total, setTotal] = useState(0);
  const [latencyData, setLatencyData] = useState<Record<string, number | null>>(
    {}
  );
  const [refreshInterval, setRefreshInterval] = useState<number | null>(10000); // 默认10秒刷新
  const [searchTerm, setSearchTerm] = useState("");
  const [displayedHosts, setDisplayedHosts] = useState<Host[]>([]);

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
      const initialLatencyData: Record<string, number | null> = {};
      hosts.forEach((host) => {
        initialLatencyData[host.host] = null;
      });
      setLatencyData(initialLatencyData);

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
    const fetchLatency = async () => {
      if (hosts?.length === 0) {
        setLatencyData({});
        return;
      }
      try {
        const hostNames = hosts.map((host) => host.host);
        if (hostNames?.length === 0) {
          setLatencyData({});
          return;
        }
        const res = await GetLatencyWithHosts(hostNames);
        const latencyHosts = res?.hosts || [];
        const latencyMap: Record<string, number | null> = {};
        latencyHosts.forEach(
          (latencyHost: { host: string; avg_latency: number | null }) => {
            latencyMap[latencyHost.host] = latencyHost.avg_latency;
          }
        );
        setLatencyData(latencyMap);
      } catch (err) {
        console.error("Error fetching latency data:", err);
      }
    };

    if (hosts.length > 0) {
      fetchLatency();
    } else {
      setLatencyData({});
    }

    let intervalId: NodeJS.Timeout | null = null;
    if (refreshInterval && hosts.length > 0) {
      intervalId = setInterval(fetchLatency, refreshInterval);
    }
    return () => {
      if (intervalId) clearInterval(intervalId);
    };
  }, [hosts, refreshInterval]);

  const handleBackendSearch = async (searchTerm: string) => {
    setPage(1); // 重置到第一页
    await fetchAndSetHosts(1, searchTerm);
  };

  return (
    <Box p="4">
      <Stack
        mx="4"
        direction="row"
        gap="4"
        align="center"
        justify="space-between"
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
      <HostList hosts={displayedHosts} latencyData={latencyData} />
      <PaginationControls
        page={page}
        total={total}
        pageSize={pageSize}
        onPageChange={setPage}
      />
    </Box>
  );
}
