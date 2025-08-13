import { Box, Text } from "@chakra-ui/react";
import {
  forwardRef,
  useCallback,
  useEffect,
  useImperativeHandle,
  useRef,
} from "react";
import AutoSizer from "react-virtualized-auto-sizer";
import { ListOnScrollProps, VariableSizeList as List } from "react-window";

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

        // 处理多行内容，保持换行符
        const content = log.raw || log.message;
        const lines = content.split("\n");

        if (log.level === "continuation") {
          return (
            <Box
              style={style}
              fontSize="sm"
              whiteSpace="pre-wrap"
              fontFamily="mono"
            >
              {lines.map((line, lineIndex) => (
                <Text key={lineIndex}>{line}</Text>
              ))}
            </Box>
          );
        }

        // 对于普通日志行，如果是多行内容，也要正确处理
        if (lines.length > 1) {
          return (
            <Box style={style}>
              <LogLine log={{ ...log, message: lines[0] }} style={{}} />
              {lines.slice(1).map((line, lineIndex) => (
                <Text
                  key={lineIndex + 1}
                  fontSize="sm"
                  whiteSpace="pre-wrap"
                  fontFamily="mono"
                  pl={4}
                >
                  {line}
                </Text>
              ))}
            </Box>
          );
        }

        return <LogLine log={log} style={style} />;
      },
      [logs]
    );

    // 添加计算行高的函数
    const getItemSize = useCallback(
      (index: number) => {
        const log = logs[index];
        if (!log) return 28;

        // 计算行数
        const lineCount = (log.raw || log.message).split("\n").length;
        // 基础高度24px + 每额外行18px
        return 28 + (lineCount - 1) * 18;
      },
      [logs]
    );

    return (
      <Box
        height="100%"
        borderWidth="1px"
        borderRadius="md"
        position="relative"
        overflow="hidden"
      >
        <Box height="100%" p="0.5rem" pb="1rem">
          <AutoSizer>
            {({ height, width }: { height: number; width: number }) => (
              <List
                ref={listRef}
                height={height}
                width={width}
                itemCount={logs.length}
                itemSize={getItemSize} // 使用动态高度函数
                overscanCount={20} // 提前渲染20行以实现平滑滚动
                onScroll={handleScroll}
                itemKey={itemKey}
                outerRef={outerDivRef}
              >
                {renderRow}
              </List>
            )}
          </AutoSizer>
        </Box>
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
            <Text fontSize="sm" color="orange.400" textAlign="center">
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
