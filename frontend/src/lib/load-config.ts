import axios from "axios";

/**
 * 从指定的 URL 加载配置
 * @param url 配置文件的 URL
 * @returns 返回原始 YAML 字符串
 */
export async function loadConfigFromUrl(url: string): Promise<string> {
  try {
    const response = await axios.get(url, { responseType: "text" });

    // 确保响应数据是字符串
    if (typeof response.data !== "string") {
      throw new Error("响应数据不是有效的字符串");
    }

    return response.data;
  } catch (error) {
    throw new Error(`加载配置失败: ${error}`);
  }
}
