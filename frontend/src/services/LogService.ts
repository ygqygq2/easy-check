import { GetLogFileContent } from "@bindings/easy-check/internal/services/appservice";

import { LogEntry } from "../types/LogTypes";
import { parseLogEntry } from "../utils/logParser";

export class LogService {
  static async getInitialLogs(
    fileName: string,
    isLatest: boolean
  ): Promise<LogEntry[]> {
    try {
      const content = await GetLogFileContent(fileName, isLatest);

      const lines = content.split("\n").filter(Boolean);
      const logEntries: LogEntry[] = [];
      let previousEntry: LogEntry | undefined = undefined;

      for (const line of lines) {
        // 判断是否为带时间戳的日志行
        const isNewLog = /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}/.test(line);

        if (isNewLog) {
          // 新日志行，先把上一条推入数组
          if (previousEntry) {
            logEntries.push(previousEntry);
          }
          const parsed = parseLogEntry(line);
          if (parsed) {
            previousEntry = parsed;
          } else {
            previousEntry = undefined;
          }
        } else if (previousEntry) {
          // 续行，拼接到上一条日志
          previousEntry.message += "\n" + line;
          previousEntry.raw += "\n" + line;
        }
      }
      // 最后一条日志别忘了加进去
      if (previousEntry) {
        logEntries.push(previousEntry);
      }

      return logEntries;
    } catch (error) {
      console.error("Error fetching log content:", error);
      throw error;
    }
  }
}
