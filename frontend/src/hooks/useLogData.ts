import { useCallback, useEffect, useRef, useState } from "react";

import { LogService } from "../services/LogService";
import { LogEntry } from "../types/LogTypes";
import { useLogWebSocket } from "./useLogWebSocket";

export function useLogData(
  fileName: string,
  isLatest: boolean,
  isRealtime: boolean,
  updateInterval: number
) {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [unreadCount, setUnreadCount] = useState<number>(0);
  const [userScrolled, setUserScrolled] = useState<boolean>(false);
  const [hasNewContent, setHasNewContent] = useState<boolean>(false);

  const lastSeenLogIdRef = useRef<string | null>(null);

  // 从WebSocket获取新日志数据
  const { newLogEntries } = useLogWebSocket(isLatest, isRealtime, updateInterval);

  // 初始加载日志
  useEffect(() => {
    const fetchLogs = async () => {
      setIsLoading(true);
      try {
        const initialLogs = await LogService.getInitialLogs(fileName, isLatest);
        setLogs(initialLogs);

        // 记录最后一条日志ID
        if (initialLogs.length > 0) {
          lastSeenLogIdRef.current = initialLogs[initialLogs.length - 1].id;
        }
      } catch (error) {
        console.error("Failed to fetch logs:", error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchLogs();
    // 重置状态
    setUnreadCount(0);
    setUserScrolled(false);
    setHasNewContent(false);
  }, [fileName, isLatest]);

  // 处理新的日志条目
  useEffect(() => {
    if (newLogEntries.length === 0) return;

    // 过滤掉已经处理过的日志
    const filteredEntries = newLogEntries.filter(
      (entry) => !logs.some((log) => log.id === entry.id)
    );

    if (filteredEntries.length === 0) return;

    if (isRealtime) {
      setLogs((currentLogs) => [...currentLogs, ...filteredEntries]);
      setHasNewContent(userScrolled); // 只有在用户已滚动时才标记有新内容
    } else {
      setUnreadCount((prev) => prev + filteredEntries.length);
    }

    // 更新最后一条日志ID
    if (filteredEntries.length > 0) {
      lastSeenLogIdRef.current = filteredEntries[filteredEntries.length - 1].id;
    }
  }, [newLogEntries, logs, isRealtime, userScrolled]);

  // 处理滚动事件
  const onScroll = useCallback(
    (scrollTop: number, scrollHeight: number, clientHeight: number) => {
      const isNearBottom = scrollHeight - scrollTop - clientHeight < 50;
      setUserScrolled(!isNearBottom);

      if (isNearBottom) {
        setHasNewContent(false); // 如果滚动到底部，重置新内容标记
      }
    },
    []
  );

  // 滚动到底部
  const scrollToBottom = useCallback(() => {
    // 注意：具体实现依赖于虚拟滚动库，这里是占位
    // 实际实现将通过ref调用滚动方法
  }, []);

  // 标记所有为已读
  const markAllAsRead = useCallback(() => {
    setUnreadCount(0);
    setHasNewContent(false);
  }, []);

  return {
    logs,
    isLoading,
    unreadCount,
    hasNewContent,
    userScrolled,
    scrollToBottom,
    markAllAsRead,
    onScroll,
  };
}
