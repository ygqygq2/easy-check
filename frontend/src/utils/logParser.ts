import md5 from "md5";

import { LogEntry, LogLevel } from "../types/LogTypes";

/**
 * 日志唯一性说明：
 * - 日志以时间戳开头，但同一时间戳可能有多条不同内容的日志。
 * - 相同时间戳不会有完全一样的日志内容。
 * - 因此，唯一 id 用 md5(`${timestamp}-${level}-${message}`)。
 */

export function parseLogEntry(line: string): LogEntry | null {
  // 检查是否是带时间戳的日志行
  const timestampMatch = line.match(/^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})/);

  if (!timestampMatch) {
    return null; // 不是有效的日志行
  }

  const level = extractLogLevel(line);
  const [timestamp, ...rest] = line.split(" ");
  const message = rest.join(" ");
  const id = md5(`${timestamp}-${level}-${message}`); // 不加Date.now，保证同一条日志唯一

  return {
    id,
    timestamp: new Date(timestamp),
    level,
    message,
    raw: line,
  };
}

function extractLogLevel(line: string): LogLevel {
  const levelRegex = /\[(info|error|warn|debug)\]/i;
  const levelMatch = line.match(levelRegex);
  return levelMatch ? (levelMatch[1].toLowerCase() as LogLevel) : "info";
}
