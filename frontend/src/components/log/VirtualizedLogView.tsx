import { Box, Text } from "@chakra-ui/react";
import {
  forwardRef,
  useCallback,
  useEffect,
  useImperativeHandle,
  useRef,
} from "react";
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
  shouldScrollToBottom?: boolean;
}

export interface VirtualizedLogViewRef {
  scrollToBottom: () => void;
}

export const VirtualizedLogView = forwardRef<
  VirtualizedLogViewRef,
  VirtualizedLogViewProps
>(
  (
    {
      logs,
      onScroll,
      hasNewContent,
      unreadCount,
      isRealtime,
      onScrollToBottom,
      userScrolled,
      shouldScrollToBottom = false,
    },
    ref
  ) => {
    const listRef = useRef<List>(null);
    const outerDivRef = useRef<HTMLDivElement>(null);

    const scrollToBottom = useCallback(() => {
      if (listRef.current && logs.length > 0) {
        listRef.current.scrollToItem(logs.length - 1, "end");
      }
    }, [logs.length]);

    // 暴露 scrollToBottom 方法给父组件
    useImperativeHandle(
      ref,
      () => ({
        scrollToBottom,
      }),
      [scrollToBottom]
    );

    // 当有新日志且处于实时模式时，自动滚动到底部
    useEffect(() => {
      if (isRealtime && logs.length > 0 && !userScrolled) {
        scrollToBottom();
      }
    }, [logs.length, isRealtime, userScrolled, scrollToBottom]);

    // 初始滚动到底部（针对最新日志查看）
    useEffect(() => {
      if (shouldScrollToBottom && logs.length > 0) {
        // 使用 requestAnimationFrame 确保渲染完成
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            scrollToBottom();
          });
        });
      }
    }, [shouldScrollToBottom, logs.length, scrollToBottom]);

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
      <Box
        height="600px"
        borderWidth="1px"
        borderRadius="md"
        position="relative"
      >
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
  }
);

VirtualizedLogView.displayName = "VirtualizedLogView";
