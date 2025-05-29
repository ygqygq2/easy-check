import {
  CheckboxCard,
  HStack,
  Progress,
  SimpleGrid,
  Stack,
} from "@chakra-ui/react";

import { Tooltip } from "@/components/ui/tooltip";
import { Host } from "@/types/host";

interface HostListProps {
  hosts: Host[];
  latencyData: Record<string, number | null>;
}

export function HostList({ hosts, latencyData }: HostListProps) {
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

  return (
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
                      {latencyData[host.host] === null
                        ? "加载中"
                        : latencyData[host.host]?.toFixed(0)}
                      ms
                    </Progress.ValueText>
                  </HStack>
                </Progress.Root>
              </CheckboxCard.Label>
            </CheckboxCard.Control>
          </CheckboxCard.Root>
        </Stack>
      ))}
    </SimpleGrid>
  );
}
