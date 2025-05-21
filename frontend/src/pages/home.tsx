"use client";
import { Checkbox, CheckboxCard, Flex, For, Progress, Stack, Text } from "@chakra-ui/react";
import { useEffect, useState } from "react";

interface HostStatus {
  id: string;
  name: string;
  latency: number; // 延迟值，单位 ms
}

export function Page() {
  const [hosts, setHosts] = useState<HostStatus[]>([]);

  useEffect(() => {
    // 替换为实际的 API 调用
    const fetchHosts = async () => {
      const data: HostStatus[] = [
        { id: "1", name: "Host A", latency: 50 },
        { id: "2", name: "Host B", latency: 150 },
        { id: "3", name: "Host C", latency: 250 },
      ];
      setHosts(data);
    };
    fetchHosts();
  }, []);

  // 根据延迟值设置进度条颜色
  const getColorScheme = (latency: number) => {
    if (latency <= 100) return "green";
    if (latency <= 200) return "yellow";
    return "red";
  };

  return (
    <>
      <Flex
        id="App"
        height="90vh"
        width="100%"
        direction="column"
        alignItems="center"
        justifyContent="flex-start"
        padding="20px"
        overflowY="auto"
      >
        {hosts.map((host) => (
          <Flex
            key={host.id}
            width="100%"
            maxWidth="600px"
            alignItems="center"
            marginBottom="10px"
            padding="10px"
            borderWidth="1px"
            borderRadius="md"
            boxShadow="sm"
          >
            <Stack maxW="320px">
              <CheckboxCard.Root defaultChecked variant="outline" colorPalette="teal">
                <CheckboxCard.HiddenInput />
                <CheckboxCard.Control>
                  <CheckboxCard.Label>Checkbox {host.name}</CheckboxCard.Label>
                  <CheckboxCard.Indicator />
                </CheckboxCard.Control>
              </CheckboxCard.Root>
            </Stack>
            <Progress.Root maxW="240px">
              <Progress.Track>
                <Progress.Range />
              </Progress.Track>
            </Progress.Root>
            <Text marginLeft="10px" width="50px" textAlign="right">
              {host.latency}ms
            </Text>
          </Flex>
        ))}
      </Flex>
    </>
  );
}
