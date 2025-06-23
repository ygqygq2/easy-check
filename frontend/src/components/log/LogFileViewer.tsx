import { Box, VStack } from "@chakra-ui/react";
import { useEffect, useState } from "react";

import { useLogData } from "../../hooks/useLogData";
import { StatusView } from "../StatusView";
import { HeaderWithActions } from "../ui/HeaderWithActions";
import { LogControlPanel } from "./LogControlPanel";
import { VirtualizedLogView } from "./VirtualizedLogView";

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

  const {
    logs,
    isLoading,
    unreadCount,
    hasNewContent,
    scrollToBottom,
    markAllAsRead,
    onScroll,
    userScrolled,
  } = useLogData(fileName, isLatest, isRealtime, updateInterval);

  useEffect(() => {
    if (logs.length > 0 && isRealtime) {
      scrollToBottom(); // 日志加载完成后滚动到底部
    }
  }, [logs, isRealtime]);

  return (
    <Box p={4}>
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
      <VStack align="stretch" gap={2}>
        {isLoading ? (
          <StatusView message="正在加载文件内容..." isLoading />
        ) : logs.length > 0 ? (
          <VirtualizedLogView
            logs={logs}
            onScroll={onScroll}
            hasNewContent={hasNewContent && isRealtime}
            unreadCount={unreadCount}
            isRealtime={isRealtime}
            onScrollToBottom={() => {
              scrollToBottom();
              markAllAsRead();
            }}
            userScrolled={userScrolled}
          />
        ) : (
          <StatusView message="文件内容为空" />
        )}
      </VStack>
    </Box>
  );
}

export default LogFileViewer;
