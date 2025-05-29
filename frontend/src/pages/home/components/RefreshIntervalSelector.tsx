import { NativeSelect, Stack, Text } from "@chakra-ui/react";

interface RefreshIntervalSelectorProps {
  refreshInterval: number | null;
  onChange: (value: number | null) => void;
}

export function RefreshIntervalSelector({
  refreshInterval,
  onChange,
}: RefreshIntervalSelectorProps) {
  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    onChange(value === "0" ? null : parseInt(value, 10) * 1000); // 转换为毫秒
  };

  return (
    <Stack direction="row" align="center" justify="flex-end">
      <Text>自动刷新</Text>
      <NativeSelect.Root size="sm" width="120px">
        <NativeSelect.Field
          value={refreshInterval ? refreshInterval / 1000 : "0"}
          onChange={handleChange}
        >
          <option value="5">5秒</option>
          <option value="10">10秒</option>
          <option value="30">30秒</option>
          <option value="0">关闭</option>
        </NativeSelect.Field>
        <NativeSelect.Indicator />
      </NativeSelect.Root>
    </Stack>
  );
}
