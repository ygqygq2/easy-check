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
        const entry = parseLogEntry(line, previousEntry);
        if (entry) {
          logEntries.push(entry);
          previousEntry = entry; // 更新上一条日志
        }
      }

      return logEntries;
    } catch (error) {
      console.error("Error fetching log content:", error);
      throw error;
    }
  }
}
