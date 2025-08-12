import { NativeSelect, Stack, Text, HStack, Button } from "@chakra-ui/react";

export interface TimeRange {
  label: string;
  minutes: number;
}

export const TIME_RANGES: TimeRange[] = [
  { label: "最近10分钟", minutes: 10 },
  { label: "最近30分钟", minutes: 30 },
  { label: "最近1小时", minutes: 60 },
  { label: "最近3小时", minutes: 180 },
  { label: "最近12小时", minutes: 720 },
  { label: "最近24小时", minutes: 1440 },
  { label: "最近2天", minutes: 2880 },
  { label: "最近7天", minutes: 10080 },
  { label: "最近30天", minutes: 43200 },
];

interface TimeRangeSelectorProps {
  selectedRange: TimeRange;
  onRangeChange: (range: TimeRange) => void;
  customRange?: { start: number; end: number } | null;
  onClearCustomRange?: () => void;
}

export default function TimeRangeSelector({
  selectedRange,
  onRangeChange,
  customRange,
  onClearCustomRange,
}: TimeRangeSelectorProps) {
  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const minutes = parseInt(e.target.value, 10);
    const range = TIME_RANGES.find((r) => r.minutes === minutes);
    if (range) {
      onRangeChange(range);
    }
  };

  return (
    <Stack direction="row" align="center" justify="flex-end">
      <Text>时间段</Text>
      {customRange ? (
        <HStack gap="2">
          <Text fontSize="xs" color="gray.600">
            {new Date(customRange.start).toLocaleString()} —{" "}
            {new Date(customRange.end).toLocaleString()}
          </Text>
          <Button size="xs" variant="outline" onClick={onClearCustomRange}>
            重置
          </Button>
        </HStack>
      ) : null}
      <NativeSelect.Root size="sm" width="140px">
        <NativeSelect.Field
          value={selectedRange.minutes}
          onChange={handleChange}
        >
          {TIME_RANGES.map((range) => (
            <option key={range.minutes} value={range.minutes}>
              {range.label}
            </option>
          ))}
        </NativeSelect.Field>
        <NativeSelect.Indicator />
      </NativeSelect.Root>
    </Stack>
  );
}
