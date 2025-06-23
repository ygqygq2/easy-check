import { useCallback, useEffect, useRef, useState } from "react";

import { LogEntry } from "../types/LogTypes";
import { parseLogEntry } from "../utils/logParser";

export function useLogWebSocket(
  isLatest: boolean,
  isRealtime: boolean,
  updateInterval: number
) {
  const [newLogEntries, setNewLogEntries] = useState<LogEntry[]>([]);
  const wsRef = useRef<WebSocket | null>(null);
  const recentMessagesRef = useRef<Set<string>>(new Set());
  const bufferRef = useRef<LogEntry[]>([]);
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  // 处理收到的日志消息
  const processLogMessages = useCallback((data: string) => {
    const lines = data.split("\n").filter((line) => line.trim() !== "");

    const newEntries: LogEntry[] = [];
    let currentEntry: LogEntry | null = null;

    for (const line of lines) {
      // 减少重复检查的严格性，只检查完整的日志行
      const timestampMatch = line.match(
        /^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})/
      );

      if (timestampMatch) {
        // 检查是否重复（只对完整日志行进行检查）
        const logKey = `${timestampMatch[1]}_${line.slice(0, 100)}`;
        if (recentMessagesRef.current.has(logKey)) {
          continue;
        }
        recentMessagesRef.current.add(logKey);

        // 如果有当前正在处理的日志，先添加到结果中
        if (currentEntry) {
          newEntries.push(currentEntry);
        }

        // 解析新的日志行
        const entry = parseLogEntry(line);
        if (entry) {
          currentEntry = entry;
        }
      } else {
        // 这是续行内容，追加到当前日志
        if (currentEntry) {
          currentEntry.message += "\n" + line;
          currentEntry.raw += "\n" + line;
        }
      }
    }

    // 处理最后一个日志条目
    if (currentEntry) {
      newEntries.push(currentEntry);
    }

    // 立即处理新日志，不使用缓冲区
    if (newEntries.length > 0) {
      setNewLogEntries((prev) => [...prev, ...newEntries]);
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

    timerRef.current = setInterval(flushBuffer, updateInterval * 1000);

    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    };
  }, [isRealtime, isLatest, updateInterval]);

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

        // 如果有日志，立即刷新缓冲区
        if (bufferRef.current.length > 0 && isRealtime) {
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
