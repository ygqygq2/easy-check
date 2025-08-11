import { useState, useCallback, useEffect } from "react";
import {
  GetFrontendConfig,
  GetConfigValue,
  GetAppInfo,
} from "@bindings/easy-check/internal/services/appservice";

// 前端配置接口，对应后端的FrontendConfig
interface FrontendConfig {
  pingInterval: number; // ping间隔时间
  globalInterval: number; // 全局间隔时间
  hostsCount: number; // 主机数量
}

// 应用信息接口，对应后端的AppInfo
interface AppInfo {
  appName: string; // 应用名称
  appVersion: string; // 应用版本
  author: string; // 作者
  copyright: string; // 版权信息
  license: string; // 许可证
  repository: string; // 代码仓库
  description: string; // 应用描述
  buildTime: string; // 构建时间
  goVersion: string; // Go版本
  platformInfo: {
    os: string; // 操作系统
    arch: string; // 架构
  };
  updateServer: string; // 更新服务器
  needsRestart: boolean; // 是否需要重启
}

// 默认配置值
const DEFAULT_PING_INTERVAL = 30;
const DEFAULT_GLOBAL_INTERVAL = 10;

export const useConfig = () => {
  const [config, setConfig] = useState<FrontendConfig | null>(null);
  const [appInfo, setAppInfo] = useState<AppInfo | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 加载前端配置
  const loadConfig = useCallback(async () => {
    if (loading) return;

    setLoading(true);
    setError(null);

    try {
      const frontendConfig = await GetFrontendConfig();
      if (frontendConfig) {
        setConfig(frontendConfig);
        console.log("Frontend config loaded:", frontendConfig);
      } else {
        throw new Error("Empty config response");
      }
    } catch (err) {
      const errorMessage = `Failed to load config: ${
        err instanceof Error ? err.message : "Unknown error"
      }`;
      console.error(errorMessage);
      setError(errorMessage);

      // 使用默认配置作为 fallback
      setConfig({
        pingInterval: DEFAULT_PING_INTERVAL,
        globalInterval: DEFAULT_GLOBAL_INTERVAL,
        hostsCount: 0,
      });
    } finally {
      setLoading(false);
    }
  }, [loading]);

  // 加载应用信息
  const loadAppInfo = useCallback(async () => {
    try {
      const info = await GetAppInfo();
      if (info) {
        setAppInfo(info);
        console.log("App info loaded:", info);
      }
    } catch (err) {
      console.error("Failed to load app info:", err);
    }
  }, []);

  // 获取特定配置值（支持YAML路径）
  const getConfigValue = useCallback(async (path: string) => {
    try {
      const value = await GetConfigValue(path);
      console.log(`Config value for ${path}:`, value);
      return value;
    } catch (err) {
      console.error(`Failed to get config value for ${path}:`, err);
      return null;
    }
  }, []);

  // 获取ping间隔时间
  const getPingInterval = useCallback(() => {
    return config?.pingInterval || DEFAULT_PING_INTERVAL;
  }, [config]);

  // 获取全局间隔时间
  const getGlobalInterval = useCallback(() => {
    return config?.globalInterval || DEFAULT_GLOBAL_INTERVAL;
  }, [config]);

  useEffect(() => {
    // 只在组件初始化时加载一次
    if (!config && !loading) {
      loadConfig();
    }
    if (!appInfo) {
      loadAppInfo();
    }
  }, []); // 空依赖数组，只在组件挂载时执行一次

  return {
    // 配置相关
    config,
    loading,
    error,
    loadConfig,
    getPingInterval,
    getGlobalInterval,
    getConfigValue,

    // 应用信息相关
    appInfo,
    loadAppInfo,
  };
};
