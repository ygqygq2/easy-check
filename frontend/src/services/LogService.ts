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

      // 解析日志文本为结构化数据
      return content
        .split("\n")
        .filter(Boolean)
        .map(parseLogEntry)
        .filter(Boolean) as LogEntry[];
    } catch (error) {
      console.error("Error fetching log content:", error);
      throw error;
    }
  }
}
