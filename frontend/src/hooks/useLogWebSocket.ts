import { useCallback, useEffect, useRef, useState } from "react";

import { LogEntry } from "../types/LogTypes";
import { parseLogEntry } from "../utils/logParser";

export function useLogWebSocket(isLatest: boolean, isRealtime: boolean) {
  const [newLogEntries, setNewLogEntries] = useState<LogEntry[]>([]);
  const wsRef = useRef<WebSocket | null>(null);
  const recentMessagesRef = useRef<Set<string>>(new Set());
  const bufferRef = useRef<LogEntry[]>([]);
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  // 处理收到的日志消息
  const processLogMessages = useCallback((data: string) => {
    const lines = data.split("\n").filter(Boolean);

    const newEntries: LogEntry[] = lines
      .map((line) => {
        // 检查是否重复
        if (recentMessagesRef.current.has(line)) {
          return null;
        }

        // 添加到最近消息集合
        recentMessagesRef.current.add(line);

        // 限制最近消息集合大小
        if (recentMessagesRef.current.size > 200) {
          const entries = Array.from(recentMessagesRef.current);
          recentMessagesRef.current = new Set(entries.slice(-100));
        }

        return parseLogEntry(line);
      })
      .filter(Boolean) as LogEntry[];

    if (newEntries.length > 0) {
      bufferRef.current = [...bufferRef.current, ...newEntries];
    }
  }, []);

  // 定期刷新日志缓冲区
  useEffect(() => {
    if (!isRealtime || !isLatest) return;

    const flushBuffer = () => {
      if (bufferRef.current.length > 0) {
        setNewLogEntries(bufferRef.current);
        bufferRef.current = [];
      }
    };

    timerRef.current = setInterval(flushBuffer, 1000);

    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    };
  }, [isRealtime, isLatest]);

  // WebSocket连接
  useEffect(() => {
    if (!isLatest) return;

    const connectWebSocket = () => {
      const ws = new WebSocket(`ws://127.0.0.1:32180/ws/logs`);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log("WebSocket connection established");
      };

      ws.onmessage = (event) => {
        processLogMessages(event.data);

        // 如果包含错误信息，立即刷新缓冲区
        if (event.data.includes("[error]") && isRealtime) {
          setNewLogEntries(bufferRef.current);
          bufferRef.current = [];
        }
      };

      ws.onclose = () => {
        console.log("WebSocket connection closed, retrying...");
        setTimeout(connectWebSocket, 5000);
      };

      ws.onerror = (error) => {
        console.error("WebSocket error:", error);
      };
    };

    connectWebSocket();

    return () => {
      wsRef.current?.close();
    };
  }, [isLatest, processLogMessages, isRealtime]);

  return { newLogEntries };
}
