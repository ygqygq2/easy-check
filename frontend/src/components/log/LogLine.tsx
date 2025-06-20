import { Box, Text } from "@chakra-ui/react";
import { memo } from "react";

import { LogEntry } from "../../types/LogTypes";

interface LogLineProps {
  log: LogEntry;
  style: React.CSSProperties;
}

export const LogLine = memo(({ log, style }: LogLineProps) => {
  // 根据日志级别设置颜色
  const getColorForLevel = (level: string): string => {
    switch (level) {
      case "error":
        return "red.600";
      case "warn":
        return "orange.600";
      case "info":
        return "gray.800";
      case "debug":
        return "gray.500";
      default:
        return "gray.800";
    }
  };

  // 日志的背景色
  const getBgColorForLevel = (level: string): string => {
    switch (level) {
      case "error":
        return "red.50";
      case "warn":
        return "orange.50";
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
      borderBottomColor="gray.100"
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
