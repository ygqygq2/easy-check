import { getSiteURL } from "./lib/get-site-url";
import { LogLevel } from "./lib/logger";

export interface Config {
  site: {
    name: string;
    description: string;
    url: string;
    version: string;
  };
  logLevel: keyof typeof LogLevel;
}

export const config = {
  site: {
    name: "Easy Check",
    description: "简单的网络检测工具",
    url: getSiteURL(),
    version: import.meta.env.VITE_SITE_VERSION || "0.0.0",
  },
  logLevel: (import.meta.env.VITE_LOG_LEVEL as keyof typeof LogLevel) || LogLevel.ALL,
} satisfies Config;
