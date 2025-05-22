"use client";
import {
  Box,
  ButtonGroup,
  CheckboxCard,
  Flex,
  HStack,
  IconButton,
  Input,
  NativeSelect,
  Pagination,
  Progress,
  SimpleGrid,
  Stack,
  Text,
} from "@chakra-ui/react";
import { Icon } from "@iconify/react";
import { useEffect, useState } from "react";

import { Tooltip } from "@/components/ui/tooltip";
import { Host } from "@/types/host";

import { GetHosts, GetLatencyWithHosts } from "../../wailsjs/go/main/App";

export function Page() {
  const pageSize = 20;
  const [page, setPage] = useState(1);
  const [hosts, setHosts] = useState<Host[]>([]);
  const [total, setTotal] = useState(0);
  const [latencyData, setLatencyData] = useState<Record<string, number | null>>({});
  const [refreshInterval, setRefreshInterval] = useState<number | null>(10000); // 默认10秒刷新

  useEffect(() => {
    const fetchHosts = async () => {
      try {
        const res = await GetHosts(page, pageSize);
        const { hosts, total } = res;
        setHosts(hosts);
        setTotal(total);

        // 初始化 latencyData 为 null
        const initialLatencyData: Record<string, number | null> = {};
        hosts.forEach((host) => {
          initialLatencyData[host.host] = null;
        });
        setLatencyData(initialLatencyData);
      } catch (err) {
        console.error("Error fetching hosts:", err);
      }
    };
    fetchHosts();
  }, [page]);

  useEffect(() => {
    const fetchLatency = async () => {
      try {
        const hostNames = hosts.map((host) => host.host);
        const res = await GetLatencyWithHosts(hostNames);
        const { hosts: latencyHosts } = res;
        const latencyMap: Record<string, number> = {};
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        latencyHosts.forEach((latencyHost: any) => {
          latencyMap[latencyHost.host] = latencyHost.avg_latency; // 使用 avg_latency
        });
        setLatencyData(latencyMap);
      } catch (err) {
        console.error("Error fetching latency data:", err);
      }
    };

    // 初次调用
    fetchLatency();

    // 根据刷新间隔设置定时器
    let interval: NodeJS.Timeout | null = null;
    if (refreshInterval) {
      interval = setInterval(fetchLatency, refreshInterval);
    }

    // 清除定时器
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [hosts, refreshInterval]);

  const getColorPalette = (latency: number) => {
    if (latency === 0) return "gray";
    if (latency <= 80) return "green";
    if (latency <= 200) return "yellow";
    return "red";
  };

  const getValue = (latency: number) => {
    if (latency <= 10) return latency / 4;
    if (latency <= 40) return latency / 4;
    if (latency <= 80) return latency / 3;
    if (latency <= 200) return latency / 3;
    return latency / 2;
  };

  const getVariant = (latency: number) => {
    if (latency === 0) return "subtle";
    return "outline";
  };

  const getDisabled = (latency: number) => {
    if (latency === 0) return true;
    return false;
  };

  const handleRefreshChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    if (value === "close") {
      setRefreshInterval(null); // 停止刷新
    } else {
      setRefreshInterval(parseInt(value, 10) * 1000); // 设置刷新间隔（毫秒）
    }
  };

  return (
    <Box p="4">
      <Stack mx="4" direction="row" gap="4" align="center">
        <Input placeholder="搜索主机名称或 IP 地址" flex="1" mr={10} />

        <Stack direction="row" align="center" flex="1" justify="flex-end">
          <Text>自动刷新</Text>
          <NativeSelect.Root size="sm" width="120px">
            <NativeSelect.Field
              placeholder="默认 10s"
              onChange={(e) => handleRefreshChange(e as unknown as React.ChangeEvent<HTMLSelectElement>)}
            >
              <option value="5">5s</option>
              <option value="10">10s</option>
              <option value="30">30s</option>
              <option value="0">关</option>
            </NativeSelect.Field>
            <NativeSelect.Indicator />
          </NativeSelect.Root>
        </Stack>
      </Stack>

      <SimpleGrid columns={{ base: 1, md: 2 }} mt="4">
        {hosts.map((host) => (
          <Stack align="center" direction="row" gap="10" px="4" key={host.host}>
            <CheckboxCard.Root
              disabled={getDisabled(latencyData[host.host] || 0)}
              variant={getVariant(latencyData[host.host] || 0)}
              colorPalette="teal"
            >
              <CheckboxCard.HiddenInput />
              <CheckboxCard.Control>
                <CheckboxCard.Indicator />
                <CheckboxCard.Label>
                  <Progress.Root
                    value={getValue(latencyData[host.host] || 0)}
                    width="100%"
                    colorPalette={getColorPalette(latencyData[host.host] || 0)}
                  >
                    <HStack gap="2">
                      <Tooltip content={host.host}>
                        <Progress.Label maxW={100} mr="2">
                          {host.host}
                        </Progress.Label>
                      </Tooltip>
                      <Progress.Track flex="1">
                        <Progress.Range />
                      </Progress.Track>
                      <Progress.ValueText>
                        {latencyData[host.host] === null ? "加载中" : latencyData[host.host]?.toFixed(0)}ms
                      </Progress.ValueText>
                    </HStack>
                  </Progress.Root>
                </CheckboxCard.Label>
              </CheckboxCard.Control>
            </CheckboxCard.Root>
          </Stack>
        ))}
      </SimpleGrid>

      <Pagination.Root
        count={Math.ceil(total / pageSize)}
        pageSize={pageSize}
        page={page}
        onPageChange={(e) => setPage(e.page)}
      >
        <Flex justify="flex-end" mt="4">
          <ButtonGroup variant="ghost" size="sm">
            <Pagination.PrevTrigger asChild>
              <IconButton>
                <Icon icon="line-md:chevron-small-left" width="24" height="24" />
              </IconButton>
            </Pagination.PrevTrigger>

            <Pagination.Items
              render={(page) => <IconButton variant={{ base: "ghost", _selected: "outline" }}>{page.value}</IconButton>}
            />

            <Pagination.NextTrigger asChild>
              <IconButton>
                <Icon icon="line-md:chevron-small-right" width="24" height="24" />
              </IconButton>
            </Pagination.NextTrigger>
          </ButtonGroup>
        </Flex>
      </Pagination.Root>
    </Box>
  );
}
