import { Box, VStack } from "@chakra-ui/react";
import { useEffect, useRef, useState } from "react";

import { useLogData } from "../../hooks/useLogData";
import { StatusView } from "../StatusView";
import { HeaderWithActions } from "../ui/HeaderWithActions";
import { LogControlPanel } from "./LogControlPanel";
import {
  VirtualizedLogView,
  VirtualizedLogViewRef,
} from "./VirtualizedLogView";

interface LogFileViewerProps {
  fileName: string;
  onClose: () => void;
  isLatest?: boolean;
}

function LogFileViewer({
  fileName,
  onClose,
  isLatest = false,
}: LogFileViewerProps) {
  const [isRealtime, setIsRealtime] = useState<boolean>(true);
  const [updateInterval, setUpdateInterval] = useState<number>(10);
  const logViewRef = useRef<VirtualizedLogViewRef>(null);

  const {
    logs,
    isLoading,
    unreadCount,
    hasNewContent,
    markAllAsRead,
    onScroll,
    userScrolled,
  } = useLogData(fileName, isLatest, isRealtime, updateInterval);

  // 当查看最新日志且日志加载完成时，自动滚动到底部
  useEffect(() => {
    if (isLatest && !isLoading && logs.length > 0) {
      // 使用多个 requestAnimationFrame 确保完全渲染
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            logViewRef.current?.scrollToBottom();
          });
        });
      });
    }
  }, [isLatest, isLoading, logs.length]);

  return (
    <Box
      p="1rem"
      height="100%"
      display="flex"
      flexDirection="column"
      overflow="hidden"
    >
      <HeaderWithActions
        title={`查看日志文件: ${fileName}`}
        actions={
          <LogControlPanel
            isRealtime={isRealtime}
            onRealtimeChange={setIsRealtime}
            updateInterval={updateInterval}
            onUpdateIntervalChange={setUpdateInterval}
            onClose={onClose}
            isLatest={isLatest}
          />
        }
      />
      <Box flex="1" mt="0.75rem" overflow="hidden" minHeight={0}>
        {isLoading ? (
          <StatusView message="正在加载文件内容..." isLoading />
        ) : logs.length > 0 ? (
          <VirtualizedLogView
            ref={logViewRef}
            logs={logs}
            onScroll={onScroll}
            hasNewContent={hasNewContent && isRealtime}
            unreadCount={unreadCount}
            isRealtime={isRealtime}
            onScrollToBottom={() => {
              markAllAsRead();
            }}
            userScrolled={userScrolled}
            shouldScrollToBottom={isLatest && !userScrolled}
          />
        ) : (
          <StatusView message="文件内容为空" />
        )}
      </Box>
    </Box>
  );
}

export default LogFileViewer;
