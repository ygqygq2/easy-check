export type LogLevel = "info" | "error" | "warn" | "debug" | "continuation";

export interface LogEntry {
  id: string;
  timestamp: Date | null;
  level: LogLevel;
  message: string;
  raw: string;
  service?: string;
  target?: string;
  isFailure?: boolean;
}
