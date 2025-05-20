import axios from "axios";
import YAML from "yaml";

export async function loadConfigFromUrl(url: string): Promise<unknown> {
  try {
    const response = await axios.get(url, { responseType: "text" });
    const config = YAML.parse(response.data);
    return config;
  } catch (error) {
    throw new Error(`加载配置失败: ${error}`);
  }
}
