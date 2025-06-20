import { Box, VStack } from "@chakra-ui/react";
import { useState } from "react";

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

  return (
    <Box p={4}>
      <HeaderWithActions
        title={`查看日志文件: ${fileName}`}
        actions={
          <LogControlPanel
            isRealtime={isRealtime}
            onRealtimeChange={setIsRealtime}
            isLatest={isLatest}
            onClose={onClose}
            updateInterval={updateInterval}
            onUpdateIntervalChange={setUpdateInterval}
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
