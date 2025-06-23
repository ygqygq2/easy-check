import md5 from "md5";

import { LogEntry, LogLevel } from "../types/LogTypes";

/**
 * 日志唯一性说明：
 * - 日志以时间戳开头，但同一时间戳可能有多条不同内容的日志。
 * - 相同时间戳不会有完全一样的日志内容。
 * - 因此，唯一 id 用 md5(`${timestamp}-${level}-${message}`)。
 */

export function parseLogEntry(
  line: string,
  previousEntry?: LogEntry
): LogEntry | null {
  // 日志格式: YYYY-MM-DD HH:mm:ss [level] message
  const timeRegex = /^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}/;
  const levelRegex = /\[(info|error|warn|debug)\]/i;

  if (!timeRegex.test(line)) {
    // 如果不以时间戳开头，附加到上一条日志
    if (previousEntry) {
      previousEntry.message += `\n${line}`;
      previousEntry.raw += `\n${line}`;
      return null; // 不生成新的日志条目
    }
    return null; // 如果没有上一条日志，忽略该行
  }

  // 提取日志级别
  const levelMatch = line.match(levelRegex);
  const level = levelMatch ? (levelMatch[1].toLowerCase() as LogLevel) : "info";

  // 解析时间戳和消息内容
  const [timestamp, ...rest] = line.split(" ");
  const message = rest.join(" ");
  const id = md5(`${timestamp}-${level}-${message}`);

  return {
    id,
    timestamp: new Date(timestamp),
    level, // 使用提取的日志级别
    message,
    raw: line,
  };
}
