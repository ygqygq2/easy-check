import {
  CheckboxCard,
  HStack,
  Progress,
  SimpleGrid,
  Stack,
  Text,
} from "@chakra-ui/react";

import { Tooltip } from "@/components/ui/tooltip";
import { Host, HostStatus, HostStatusMap } from "@/types/host";

interface HostListProps {
  hosts: Host[];
  statusData: HostStatusMap;
  selectedHosts: string[];
  onToggleHost?: (host: string) => void;
}

export function HostList({
  hosts,
  statusData,
  selectedHosts,
  onToggleHost,
}: HostListProps) {
  // 固定主机名显示宽度，使进度条起点对齐
  const labelWidth = { base: "100px", md: "120px" };

  const getColorPalette = (latency: number | null) => {
    if (latency === null) return "gray";
    if (latency <= 80) return "green";
    if (latency <= 200) return "yellow";
    return "red";
  };

  const getValue = (latency: number | null) => {
    if (latency === null) return 0;
    if (latency <= 10) return latency / 4;
    if (latency <= 40) return latency / 4;
    if (latency <= 80) return latency / 3;
    if (latency <= 200) return latency / 3;
    return latency / 2;
  };

  const getVariant = (latency: number | null) => {
    if (latency === null) return "subtle";
    return "outline";
  };

  const getFontColor = (status: HostStatus | null) => {
    if (status?.status === "ALERT") {
      return "red";
    }
    return undefined;
  };

  const ellipsisStyle = {
    overflow: "hidden",
    textOverflow: "ellipsis",
    whiteSpace: "nowrap",
  } as const;

  return (
    <SimpleGrid columns={{ base: 1, md: 2 }} mt="4">
      {hosts.map((host) => {
        const hostStatus = statusData.get(host.host) || null;
        const latency: number | null =
          hostStatus && typeof hostStatus.latency === "number"
            ? hostStatus.latency
            : null;
        const isAlert = hostStatus?.status === "ALERT";
        const isDisabled = latency === null || isAlert;
        const isSelected = selectedHosts.includes(host.host);

        return (
          <Stack align="center" direction="row" gap="10" px="2" key={host.host}>
            <CheckboxCard.Root
              disabled={isDisabled}
              variant={isSelected ? "solid" : getVariant(latency)}
              colorPalette={isSelected ? "teal" : "teal"}
              checked={isSelected}
              onCheckedChange={(details) => {
                if (isDisabled) return;
                console.log(
                  "CheckboxCard onCheckedChange:",
                  details,
                  host.host
                );
                onToggleHost?.(host.host);
              }}
            >
              <CheckboxCard.HiddenInput />
              <CheckboxCard.Control>
                <CheckboxCard.Indicator />
                <CheckboxCard.Label>
                  {isDisabled ? (
                    <HStack gap="2" w="100%" justifyContent="space-between">
                      <Tooltip content={host.host}>
                        <Text
                          w={labelWidth}
                          mr="2"
                          color={getFontColor(hostStatus)}
                          style={ellipsisStyle}
                        >
                          {host.host}
                        </Text>
                      </Tooltip>
                      <Text color={isAlert ? "red" : "gray.500"}>
                        {isAlert ? "不可用" : "加载中"}
                      </Text>
                    </HStack>
                  ) : (
                    <Progress.Root
                      value={getValue(latency)}
                      width="100%"
                      colorPalette={getColorPalette(latency)}
                    >
                      <HStack gap="2">
                        <Tooltip content={host.host}>
                          <Progress.Label
                            w={labelWidth}
                            mr="2"
                            color={getFontColor(hostStatus)}
                            style={ellipsisStyle}
                          >
                            {host.host}
                          </Progress.Label>
                        </Tooltip>
                        <Progress.Track flex="1">
                          <Progress.Range />
                        </Progress.Track>
                        <Progress.ValueText>
                          {latency === null
                            ? "加载中"
                            : `${latency.toFixed(0)} ms`}
                        </Progress.ValueText>
                      </HStack>
                    </Progress.Root>
                  )}
                </CheckboxCard.Label>
              </CheckboxCard.Control>
            </CheckboxCard.Root>
          </Stack>
        );
      })}
    </SimpleGrid>
  );
}
