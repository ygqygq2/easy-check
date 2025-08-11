import { useColorModeValue } from "@/components/ui/color-mode";

interface SimpleTooltipProps {
  active?: boolean;
  payload?: any[];
  label?: string | number;
}

/**
 * Recharts图表的简化工具提示组件
 * 显示延迟和丢包率信息
 */
export const SimpleTooltip = ({
  active,
  payload,
  label,
}: SimpleTooltipProps) => {
  const tooltipBg = useColorModeValue(
    "rgba(255,255,255,0.95)",
    "rgba(45,55,72,0.95)"
  );
  const tooltipBorder = useColorModeValue("#e2e8f0", "#4a5568");

  if (!active || !payload || !payload.length) return null;

  // 获取当前时间点的原始数据
  const currentData = payload[0]?.payload;

  // 只显示平均延迟，对于丢包率我们直接从原始数据中获取
  const avgItems = payload.filter((item: any) => {
    const dataKey = item.dataKey as string;
    return dataKey.includes(":avg");
  });

  return (
    <div
      style={{
        fontSize: "12px",
        backgroundColor: tooltipBg,
        border: `1px solid ${tooltipBorder}`,
        borderRadius: "4px",
        boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
        padding: "8px",
      }}
    >
      <p style={{ margin: "0 0 4px 0" }}>
        {new Date(Number(label)).toLocaleString("zh-CN", { hour12: false })}
      </p>
      {avgItems.map((item: any, index: number) => {
        const dataKey = item.dataKey as string;
        const hostName = dataKey.split(":")[0];
        const avgValue = Number(item.value);

        // 从原始数据中获取丢包率
        const lossKey = `${hostName}:loss`;
        const lossValue =
          currentData && typeof currentData[lossKey] === "number"
            ? currentData[lossKey]
            : 0;

        return (
          <div key={index}>
            <p
              style={{
                margin: "2px 0",
                color: item.color,
                display: "flex",
                justifyContent: "space-between",
              }}
            >
              <span>{hostName} 平均延迟:</span>
              <span style={{ marginLeft: "8px", fontWeight: "bold" }}>
                {avgValue.toFixed(2)}ms
              </span>
            </p>
            <p
              style={{
                margin: "2px 0",
                color: item.color,
                display: "flex",
                justifyContent: "space-between",
              }}
            >
              <span>{hostName} 丢包率:</span>
              <span style={{ marginLeft: "8px", fontWeight: "bold" }}>
                {lossValue.toFixed(2)}%
              </span>
            </p>
          </div>
        );
      })}
    </div>
  );
};

export default SimpleTooltip;
