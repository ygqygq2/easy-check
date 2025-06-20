import md5 from "md5";

import { LogEntry, LogLevel } from "../types/LogTypes";

/**
 * 日志唯一性说明：
 * - 日志以时间戳开头，但同一时间戳可能有多条不同内容的日志。
 * - 相同时间戳不会有完全一样的日志内容。
 * - 因此，唯一 id 用 md5(`${timestamp}-${level}-${message}`)。
 */

export function parseLogEntry(line: string): LogEntry | null {
  // 日志格式: YYYY-MM-DD HH:mm:ss [level] message
  const timeRegex =
    /^(\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2})\s+\[(info|error|warn|debug)\]\s+(.+)$/i;
  const match = line.match(timeRegex);

  if (!match) {
    // 如果不符合格式，可能是多行日志的一部分
    return {
      id: md5(`cont-${line}`),
      timestamp: null,
      level: "continuation" as LogLevel,
      message: line,
      raw: line,
    };
  }

  const [, timestamp, levelStr, message] = match;
  const level = levelStr.toLowerCase() as LogLevel;
  // 用 md5 生成唯一 id
  const id = md5(`${timestamp}-${level}-${message}`);

  // 检查Ping相关日志的特殊格式
  if (message.includes("Ping to") && message.includes("failed")) {
    return {
      id,
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
      id,
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
    id,
    timestamp: new Date(timestamp),
    level,
    message,
    raw: line,
  };
}
