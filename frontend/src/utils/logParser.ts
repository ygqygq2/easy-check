import { v4 as uuidv4 } from "uuid";

import { LogEntry, LogLevel } from "../types/LogTypes";

export function parseLogEntry(line: string): LogEntry | null {
  // 日志格式: YYYY-MM-DD HH:mm:ss [level] message
  const timeRegex =
    /^(\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2})\s+\[(info|error|warn|debug)\]\s+(.+)$/i;
  const match = line.match(timeRegex);

  if (!match) {
    // 如果不符合格式，可能是多行日志的一部分
    return {
      id: uuidv4(),
      timestamp: null,
      level: "continuation" as LogLevel,
      message: line,
      raw: line,
    };
  }

  const [, timestamp, levelStr, message] = match;
  const level = levelStr.toLowerCase() as LogLevel;

  // 检查Ping相关日志的特殊格式
  if (message.includes("Ping to") && message.includes("failed")) {
    return {
      id: uuidv4(),
      timestamp: new Date(timestamp),
      level: "error",
      message,
      raw: line,
      service: message.match(/\[([^\]]+)\]/)?.[1] || "",
      target: message.match(/\] ([^\s]+)/)?.[1] || "",
      isFailure: true,
    };
  }

  if (message.includes("Ping to") && message.includes("succeeded")) {
    return {
      id: uuidv4(),
      timestamp: new Date(timestamp),
      level: "info",
      message,
      raw: line,
      service: message.match(/\[([^\]]+)\]/)?.[1] || "",
      target: message.match(/\] ([^\s]+)/)?.[1] || "",
      isFailure: false,
    };
  }

  return {
    id: uuidv4(),
    timestamp: new Date(timestamp),
    level,
    message,
    raw: line,
  };
}
