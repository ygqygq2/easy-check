// SmokePing风格丢包率颜色配置 - 统一定义
export const PACKET_LOSS_COLORS = {
  0: "#00ff00", // 绿色：0% 丢包
  1: "#ffff00", // 黄色：≤1% 轻微丢包
  5: "#ffa500", // 橙色：1-5% 中等丢包
  20: "#ff6600", // 深橙色：5-20% 较高丢包
  99: "#ff69b4", // 粉红色：20-99% 严重丢包
  100: "#ff0000", // 红色：100% 完全丢包
} as const;

/**
 * 根据丢包率获取SmokePing风格的颜色
 * @param lossRate 丢包率百分比 (0-100)
 * @param baseColor 无丢包时使用的基础颜色
 * @returns 对应丢包率的颜色值
 */
export function getPacketLossColor(
  lossRate: number,
  baseColor: string
): string {
  if (lossRate === 0) return baseColor; // 无丢包：使用主延迟线颜色
  if (lossRate <= 1) return PACKET_LOSS_COLORS[1]; // 黄色：轻微丢包 (≤1%)
  if (lossRate <= 5) return PACKET_LOSS_COLORS[5]; // 橙色：中等丢包 (1-5%)
  if (lossRate <= 20) return PACKET_LOSS_COLORS[20]; // 深橙色：较高丢包 (5-20%)
  if (lossRate < 100) return PACKET_LOSS_COLORS[99]; // 粉红色：严重丢包 (20-99%)
  return PACKET_LOSS_COLORS[100]; // 红色：完全丢包 (100%)
}

/**
 * 丢包率颜色图例配置
 * 用于在UI中显示颜色说明
 */
export const PACKET_LOSS_LEGEND = [
  { color: PACKET_LOSS_COLORS[0], label: "0%" },
  { color: PACKET_LOSS_COLORS[1], label: "≤1%" },
  { color: PACKET_LOSS_COLORS[5], label: "1-5%" },
  { color: PACKET_LOSS_COLORS[20], label: "5-20%" },
  { color: PACKET_LOSS_COLORS[99], label: "20-99%" },
  { color: PACKET_LOSS_COLORS[100], label: "100%" },
] as const;
