import { Box, Text } from "@chakra-ui/react";
import { useCallback, useEffect, useRef } from "react";
import AutoSizer from "react-virtualized-auto-sizer";
import { FixedSizeList as List, ListOnScrollProps } from "react-window";

import { LogEntry } from "../../types/LogTypes";
import { LogLine } from "./LogLine";
import { NewContentIndicator } from "./NewContentIndicator";

interface VirtualizedLogViewProps {
  logs: LogEntry[];
  onScroll: (
    scrollTop: number,
    scrollHeight: number,
    clientHeight: number
  ) => void;
  hasNewContent: boolean;
  unreadCount: number;
  isRealtime: boolean;
  onScrollToBottom: () => void;
  userScrolled: boolean;
}

export const VirtualizedLogView = ({
  logs,
  onScroll,
  hasNewContent,
  unreadCount,
  isRealtime,
  onScrollToBottom,
  userScrolled,
}: VirtualizedLogViewProps) => {
  const listRef = useRef<List>(null);
  const outerDivRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = useCallback(() => {
    if (listRef.current) {
      listRef.current.scrollToItem(logs.length - 1, "end"); // 滚动到最后一项
    }
  }, [logs]);

  useEffect(() => {
    if (isRealtime) {
      scrollToBottom(); // 实时更新时滚动到底部
    }
  }, [logs, isRealtime, scrollToBottom]);

  const handleScroll = useCallback(
    (props: ListOnScrollProps) => {
      const { scrollOffset, scrollUpdateWasRequested } = props;
      if (!scrollUpdateWasRequested) {
        const element = outerDivRef.current;
        if (element) {
          const { scrollHeight, clientHeight } = element;
          onScroll(scrollOffset, scrollHeight, clientHeight);
        }
      }
    },
    [onScroll]
  );

  const itemKey = useCallback(
    (index: number) => {
      return logs[index]?.id || index;
    },
    [logs]
  );

  const renderRow = useCallback(
    ({ index, style }: { index: number; style: React.CSSProperties }) => {
      const log = logs[index];
      if (log.level === "continuation") {
        return (
          <Text style={style} fontSize="sm">
            {log.message}
          </Text>
        );
      }
      return <LogLine log={log} style={style} />;
    },
    [logs]
  );

  return (
    <Box height="600px" borderWidth="1px" borderRadius="md" position="relative">
      <AutoSizer>
        {({ height, width }: { height: number; width: number }) => (
          <List
            ref={listRef}
            height={height}
            width={width}
            itemCount={logs.length}
            itemSize={24} // 每行大约24px高
            overscanCount={20} // 提前渲染20行以实现平滑滚动
            onScroll={handleScroll}
            itemKey={itemKey}
            outerRef={outerDivRef}
          >
            {renderRow}
          </List>
        )}
      </AutoSizer>
      {!isRealtime && unreadCount > 0 && (
        <Box
          mt={2}
          p={2}
          borderRadius="md"
          position="absolute"
          bottom={4}
          left={4}
          right={4}
        >
          <Text fontSize="sm" color="orange.700" textAlign="center">
            有 {unreadCount} 条新日志未显示
          </Text>
        </Box>
      )}
      {userScrolled && hasNewContent && (
        <NewContentIndicator onScrollToBottom={onScrollToBottom} />
      )}
    </Box>
  );
};
