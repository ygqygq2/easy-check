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

  const color = getColorForLevel(log.level);

  return (
    <Box
      style={style}
      px={2}
      py={0.5}
      bg={colorMode === "dark" ? "gray.900" : "gray.50"}
      borderBottomWidth="1px"
      fontSize="sm"
    >
      <Text
        color={color}
        fontFamily="monospace"
        whiteSpace="pre-wrap" // 确保换行符被正确渲染
        wordBreak="break-word"
      >
        {log.raw} {/* 使用 log.raw 展示原始日志内容 */}
      </Text>
    </Box>
  );
});

LogLine.displayName = "LogLine";
