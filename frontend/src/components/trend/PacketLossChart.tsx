import { memo } from "react";
import { Box, HStack, Text, Wrap, WrapItem } from "@chakra-ui/react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { SeriesPoint } from "@/types/series";

interface Props {
  data: SeriesPoint[];
  selectedHosts: string[];
  // host => color
  colorMap: Record<string, string>;
}

const PacketLossChart = memo(function PacketLossChart({
  data,
  selectedHosts,
  colorMap,
}: Props) {
  return (
    <Box w="100%" h="100%">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart
          data={data}
          margin={{ top: 2, right: 16, bottom: 2, left: 0 }}
        >
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis
            dataKey="ts"
            type="number"
            domain={["dataMin", "dataMax"]}
            tickFormatter={(v) => new Date(v).toLocaleTimeString()}
          />
          <YAxis unit=" %" domain={[0, 100]} />
          <Tooltip
            labelFormatter={(l) => new Date(Number(l)).toLocaleTimeString()}
          />
          {selectedHosts.map((h) => (
            <Line
              key={`${h}-loss`}
              type="monotone"
              dataKey={`${h}:loss`}
              stroke={colorMap[h]}
              dot={false}
              strokeWidth={2}
              isAnimationActive={false}
              connectNulls
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
      {selectedHosts.length > 0 && (
        <Box mt="1" px="2">
          <Wrap gap="12px">
            {selectedHosts.map((h) => (
              <WrapItem key={`legend-loss-${h}`}>
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
        </Box>
      )}
    </Box>
  );
});

export default PacketLossChart;
