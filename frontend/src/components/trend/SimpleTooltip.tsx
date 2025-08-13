import { useColorModeValue } from "@/components/ui/color-mode";
import { SeriesPoint } from "@/types/series";

interface SimpleTooltipProps {
  active?: boolean;
  payload?: Array<{ payload?: SeriesPoint }>;
  label?: string | number;
  // 完整数据与选中的主机，用于在当前行无值时，回退到最近的有效行（Grafana 风格）
  data?: SeriesPoint[];
  selectedHosts?: string[];
}

/**
 * Recharts图表的简化工具提示组件
 * 显示延迟和丢包率信息
 * 优化版本：提供更精确的数据查找和更好的用户体验
 */
export const SimpleTooltip = ({
  active,
  payload,
  label,
  data,
  selectedHosts,
}: SimpleTooltipProps) => {
  const tooltipBg = useColorModeValue(
    "rgba(255,255,255,0.95)",
    "rgba(45,55,72,0.95)"
  );
  const tooltipBorder = useColorModeValue("#e2e8f0", "#4a5568");

  if (!active || label == null) return null;

  // 改进的数据点查找逻辑
  const findBestMatchingPoint = (targetTs: number): SeriesPoint | null => {
    if (!Array.isArray(data) || data.length === 0) return null;

    // 首先尝试精确匹配
    const exactMatch = data.find((point: SeriesPoint) => point.ts === targetTs);
    if (exactMatch) return exactMatch;

    // 如果没有精确匹配，找到最近的点
    let bestMatch = data[0];
    let bestDiff = Math.abs(data[0].ts - targetTs);

    // 使用二分查找优化性能
    let left = 0;
    let right = data.length - 1;

    while (left <= right) {
      const mid = Math.floor((left + right) / 2);
      const midPoint = data[mid];
      const diff = Math.abs(midPoint.ts - targetTs);

      if (diff < bestDiff) {
        bestDiff = diff;
        bestMatch = midPoint;
      }

      if (midPoint.ts < targetTs) {
        left = mid + 1;
      } else {
        right = mid - 1;
      }
    }

    // 检查相邻的几个点以确保找到最优匹配
    const candidateIndices = [
      left - 1,
      left,
      right,
      right + 1,
      left + 1,
      right - 1,
    ]
      .filter((idx) => idx >= 0 && idx < data.length)
      .filter((idx, index, arr) => arr.indexOf(idx) === index); // 去重

    for (const idx of candidateIndices) {
      const point = data[idx];
      const diff = Math.abs(point.ts - targetTs);
      if (diff < bestDiff) {
        bestDiff = diff;
        bestMatch = point;
      }
    }

    return bestMatch;
  };

  // 优先使用当前 payload；若当前行没有任何 avg 值，则回退到最近的有效行
  let row: SeriesPoint | undefined =
    payload && payload.length ? payload[0]?.payload : undefined;
  const hasValidData = (r: SeriesPoint | undefined) =>
    !!selectedHosts?.some(
      (h) =>
        typeof (r as unknown as Record<string, unknown>)?.[`${h}:avg`] ===
        "number"
    );

  let snapped = false;
  const targetTs = Number(label);

  if (!row || (selectedHosts && !hasValidData(row))) {
    const bestMatch = findBestMatchingPoint(targetTs);

    if (bestMatch && (!selectedHosts || hasValidData(bestMatch))) {
      row = bestMatch;
      // 只有当时间戳差异超过一定阈值时才显示"近似"标记
      const timeDiff = Math.abs(bestMatch.ts - targetTs);
      snapped = timeDiff > 1000; // 超过1秒显示近似
    }
  }

  if (!row) return null;

  // 以 row 为准组装展示项
  const avgItems = (selectedHosts || [])
    .map((host) => {
      if (!row) return null;
      const rowData = row as unknown as Record<string, unknown>;
      const avg = rowData[`${host}:avg`];
      if (typeof avg !== "number") return null;
      const loss =
        typeof rowData[`${host}:loss`] === "number"
          ? rowData[`${host}:loss`]
          : 0;
      const range = rowData[`${host}:range`];
      return { host, avg, loss, range };
    })
    .filter(Boolean) as {
    host: string;
    avg: number;
    loss: number;
    range?: [number, number];
  }[];

  if (avgItems.length === 0) return null;

  return (
    <div
      style={{
        fontSize: "12px",
        backgroundColor: tooltipBg,
        border: `1px solid ${tooltipBorder}`,
        borderRadius: "4px",
        boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
        padding: "8px",
        minWidth: "160px",
        maxWidth: "300px",
      }}
    >
      <p
        style={{
          margin: "0 0 6px 0",
          fontWeight: "bold",
          borderBottom: `1px solid ${tooltipBorder}`,
          paddingBottom: "4px",
        }}
      >
        {new Date(Number(row?.ts ?? label)).toLocaleString("zh-CN", {
          hour12: false,
          year: "numeric",
          month: "2-digit",
          day: "2-digit",
          hour: "2-digit",
          minute: "2-digit",
          second: "2-digit",
        })}
        {snapped && (
          <span style={{ color: "#666", fontSize: "10px" }}> (近似)</span>
        )}
      </p>
      {avgItems.map((item, index) => (
        <div
          key={index}
          style={{ marginBottom: index < avgItems.length - 1 ? "6px" : "0" }}
        >
          <div
            style={{
              fontWeight: "bold",
              color: "#333",
              marginBottom: "2px",
              fontSize: "11px",
            }}
          >
            {item.host}
          </div>
          <p
            style={{
              margin: "2px 0",
              display: "flex",
              justifyContent: "space-between",
              fontSize: "11px",
            }}
          >
            <span>平均延迟:</span>
            <span style={{ marginLeft: "8px", fontWeight: "bold" }}>
              {item.avg.toFixed(2)}ms
            </span>
          </p>
          {item.range &&
            Array.isArray(item.range) &&
            item.range.length === 2 && (
              <p
                style={{
                  margin: "2px 0",
                  display: "flex",
                  justifyContent: "space-between",
                  fontSize: "11px",
                  color: "#666",
                }}
              >
                <span>延迟范围:</span>
                <span style={{ marginLeft: "8px" }}>
                  {item.range[0].toFixed(1)} - {item.range[1].toFixed(1)}ms
                </span>
              </p>
            )}
          <p
            style={{
              margin: "2px 0",
              display: "flex",
              justifyContent: "space-between",
              fontSize: "11px",
            }}
          >
            <span>丢包率:</span>
            <span
              style={{
                marginLeft: "8px",
                fontWeight: "bold",
                color: item.loss > 0 ? "#e53e3e" : "#38a169",
              }}
            >
              {item.loss.toFixed(2)}%
            </span>
          </p>
        </div>
      ))}
    </div>
  );
};

export default SimpleTooltip;
