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
  defaultYamlUrl: string;
}

export const config = {
  site: {
    name: "Easy Check",
    description: "简单网络检测工具",
    url: getSiteURL(),
    version: import.meta.env.VITE_SITE_VERSION || "0.0.0",
  },
  logLevel: (import.meta.env.VITE_LOG_LEVEL as keyof typeof LogLevel) || LogLevel.ALL,
  defaultYamlUrl: "https://raw.githubusercontent.com/ygqygq2/easy-check/refs/heads/main/configs/config.yaml",
} satisfies Config;
