import { Box, Text } from "@chakra-ui/react";
import { memo } from "react";

import { LogEntry } from "../../types/LogTypes";
import { useColorMode } from "../ui/color-mode";

interface LogLineProps {
  log: LogEntry;
  style: React.CSSProperties;
}

export const LogLine = memo(({ log, style }: LogLineProps) => {
  const { colorMode } = useColorMode();

  // 根据日志级别设置颜色
  const getColorForLevel = (level: string): string => {
    switch (level) {
      case "error":
        return colorMode === "dark" ? "red.300" : "red.600";
      case "warn":
        return colorMode === "dark" ? "orange.300" : "orange.600";
      case "info":
        return colorMode === "dark" ? "gray.300" : "gray.800";
      case "debug":
        return colorMode === "dark" ? "gray.400" : "gray.500";
      default:
        return "";
    }
  };

  // 日志的背景色
  const getBgColorForLevel = (level: string): string => {
    switch (level) {
      case "error":
        return colorMode === "dark" ? "red.900" : "red.50";
      case "warn":
        return colorMode === "dark" ? "orange.900" : "orange.50";
      default:
        return "transparent";
    }
  };

  const color = getColorForLevel(log.level);
  const bgColor = getBgColorForLevel(log.level);

  return (
    <Box
      style={style}
      px={2}
      py={0.5}
      bg={bgColor}
      borderBottomWidth="1px"
      fontSize="sm"
    >
      <Text
        color={color}
        fontFamily="monospace"
        whiteSpace="pre-wrap"
        wordBreak="break-all"
      >
        {log.raw}
      </Text>
    </Box>
  );
});

LogLine.displayName = "LogLine";
