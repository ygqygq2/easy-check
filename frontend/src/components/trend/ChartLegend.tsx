import { Box, HStack, Text, Wrap, WrapItem } from "@chakra-ui/react";

import { PACKET_LOSS_LEGEND } from "./smokeping-colors";

interface ChartLegendProps {
  selectedHosts: string[];
  colorMap: Record<string, string>;
}

/**
 * 图表图例组件
 * 显示主机图例和丢包率颜色说明
 */
export const ChartLegend = ({ selectedHosts, colorMap }: ChartLegendProps) => {
  if (selectedHosts.length === 0) return null;

  return (
    <Box mt="1" px="2" flexShrink={0}>
      {/* 主机图例 */}
      <Wrap gap="12px" mb="2">
        {selectedHosts.map((h) => (
          <WrapItem key={`legend-${h}`}>
            <HStack gap="1.5">
              <Box
                w="10px"
                h="10px"
                borderRadius="full"
                bg={colorMap[h]}
                boxShadow="inset 0 0 0 1px rgba(0,0,0,0.25)"
              />
              <Text fontSize="xs" color="gray.600">
                {h}
              </Text>
            </HStack>
          </WrapItem>
        ))}
      </Wrap>

      {/* 丢包率颜色图例 */}
      <HStack gap="4" justify="center" fontSize="xs" color="gray.600">
        <Text fontWeight="medium">丢包率颜色:</Text>
        <HStack gap="3">
          {PACKET_LOSS_LEGEND.map(({ color, label }, index) => (
            <HStack key={index} gap="1">
              <Box w="8px" h="8px" bg={color} />
              <Text>{label}</Text>
            </HStack>
          ))}
        </HStack>
      </HStack>
    </Box>
  );
};

export default ChartLegend;
