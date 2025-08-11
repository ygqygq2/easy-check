"use client";
import { GetHosts } from "@bindings/easy-check/internal/services/appservice";
import { Box, Grid, GridItem, Stack } from "@chakra-ui/react";
import { useEffect, useState } from "react";

import TrendPanel from "@/components/trend/TrendPanel";
import { toaster } from "@/components/ui/toaster";
import { Host } from "@/types/host";
import { useHistoryData } from "@/hooks/useHistoryData";
import { useHostStatusRefresh } from "@/hooks/useHostStatusRefresh";
import { useHostSelection } from "@/hooks/useHostSelection";

import { PaginationControls } from "../components/PaginationControls";
import { HostList } from "./home/components/HostList";
import { RefreshIntervalSelector } from "./home/components/RefreshIntervalSelector";
import { SearchBar } from "./home/components/SearchBar";

export function Page() {
  const pageSize = 10;

  const [page, setPage] = useState(1);
  const [hosts, setHosts] = useState<Host[]>([]);
  const [total, setTotal] = useState(0);
  const [refreshInterval, setRefreshInterval] = useState<number | null>(10000);
  const [searchTerm, setSearchTerm] = useState("");
  const [displayedHosts, setDisplayedHosts] = useState<Host[]>([]);

  // 使用自定义 hooks 管理历史数据和主机状态
  const { historyMap, addDataPoint, loadHistoryForHost, fillMissingData } =
    useHistoryData();

  // 首先定义 selectedHosts
  const {
    selectedHosts,
    handleHostSelection,
    clearSelection,
    updateStatusData,
  } = useHostSelection(
    new Map(), // 初始状态，将在 statusData 获取后更新
    async (hostName) => {
      // 当选中主机时，智能补全数据：先从缓存拿，再补全缺失的部分
      console.log(`Host selected: ${hostName}, checking data completeness...`);

      const existingData = historyMap[hostName] || [];
      if (existingData.length === 0) {
        // 完全没有数据，加载完整历史
        console.log(`No cached data for ${hostName}, loading full history...`);
        await loadHistoryForHost(hostName);
      } else {
        // 有缓存数据，智能补全缺失的时间段
        console.log(
          `Found ${existingData.length} cached points for ${hostName}, filling missing data...`
        );
        await fillMissingData(hostName);
      }
    },
    (hostName) => {
      console.log(`Host deselected: ${hostName}`);
      // 取消选中时不清除历史数据，保留以备后续使用
    }
  );

  // 获取所有主机的状态和自动刷新
  const { statusData } = useHostStatusRefresh(
    hosts,
    refreshInterval,
    selectedHosts,
    addDataPoint
  );

  // 更新主机选择 hook 的状态数据
  useEffect(() => {
    if (statusData && statusData.size > 0) {
      updateStatusData(statusData);
    }
  }, [statusData, updateStatusData]);

  // 获取主机列表
  const fetchAndSetHosts = async (page: number, searchTerm: string) => {
    try {
      const res = (await GetHosts(page, pageSize, searchTerm)) || {
        hosts: [],
        total: 0,
      };
      const { hosts = [], total = 0 } = res;

      setHosts(hosts);
      setTotal(total);

      console.log(
        `Loaded ${hosts.length} hosts, total: ${total}, page: ${page}, pageSize: ${pageSize}`
      );
    } catch (err) {
      console.error("Error fetching hosts:", err);
    }
  };

  // 初始加载、翻页和搜索
  useEffect(() => {
    fetchAndSetHosts(page, searchTerm);
  }, [page, searchTerm]);

  // 搜索时重置页码
  useEffect(() => {
    if (searchTerm) {
      setPage(1); // 搜索时重置到第一页
      clearSelection();
    }
  }, [searchTerm, clearSelection]);

  // 直接显示后端返回的数据
  useEffect(() => {
    setDisplayedHosts(hosts || []);
  }, [hosts]);

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
            onSearch={() => fetchAndSetHosts(1, searchTerm)} // 搜索时从第一页开始
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
          onToggleHost={handleHostSelection}
        />
        <PaginationControls
          page={page}
          total={total} // 始终使用后端返回的总数
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
