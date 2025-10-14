import {
  Box,
  Button,
  Flex,
  HStack,
  IconButton,
  Input,
  Popover,
  Stack,
  Text,
} from "@chakra-ui/react";
import { Icon } from "@iconify/react";
import React, { useMemo, useRef, useState } from "react";

export type UnixMs = number;

export interface TimeRange {
  from: UnixMs;
  to: UnixMs;
}

type Unit = "m" | "h" | "d";

export interface QuickRange {
  label: string;
  amount: number;
  unit: Unit;
}

export interface TimeRangePickerProps {
  value?: TimeRange | null;
  onChange?: (range: TimeRange) => void;
  onApply?: (range: TimeRange) => void;
  quickRanges?: QuickRange[];
  placement?: "bottom-start" | "bottom" | "bottom-end" | "right-start";
  size?: "sm" | "md" | "lg";
  buttonLabel?: string;
}

const defaultQuickRanges: QuickRange[] = [
  { label: "最近5分钟", amount: 5, unit: "m" },
  { label: "最近15分钟", amount: 15, unit: "m" },
  { label: "最近30分钟", amount: 30, unit: "m" },
  { label: "最近1小时", amount: 1, unit: "h" },
  { label: "最近3小时", amount: 3, unit: "h" },
  { label: "最近6小时", amount: 6, unit: "h" },
  { label: "最近12小时", amount: 12, unit: "h" },
  { label: "最近24小时", amount: 24, unit: "h" },
  { label: "最近2天", amount: 2, unit: "d" },
];

function addRelative(now: number, amount: number, unit: Unit) {
  const factor = unit === "m" ? 60_000 : unit === "h" ? 3_600_000 : 86_400_000;
  return now - amount * factor;
}

// datetime-local helpers
function toLocalInputValue(ts: number) {
  const d = new Date(ts);
  const pad = (n: number) => String(n).padStart(2, "0");
  const yyyy = d.getFullYear();
  const mm = pad(d.getMonth() + 1);
  const dd = pad(d.getDate());
  const hh = pad(d.getHours());
  const mi = pad(d.getMinutes());
  return `${yyyy}-${mm}-${dd}T${hh}:${mi}`;
}
function fromLocalInputValue(v: string) {
  const ms = new Date(v).getTime();
  return Number.isFinite(ms) ? ms : Date.now();
}

// editable string helpers (YYYY-MM-DD HH:mm)
function toEditString(ts: number) {
  const d = new Date(ts);
  const pad = (n: number) => String(n).padStart(2, "0");
  const yyyy = d.getFullYear();
  const mm = pad(d.getMonth() + 1);
  const dd = pad(d.getDate());
  const hh = pad(d.getHours());
  const mi = pad(d.getMinutes());
  return `${yyyy}-${mm}-${dd} ${hh}:${mi}`;
}
function parseEditString(s: string): number | null {
  const raw = s
    .trim()
    .replace(/[./]/g, "-")
    .replace(/，/g, ",")
    .replace(/\s+/g, " ");
  const m = raw.match(/^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2})$/);
  if (m) {
    const [, Y, Mo, D, H, Mi] = m;
    const d = new Date();
    d.setFullYear(Number(Y));
    d.setMonth(Number(Mo) - 1);
    d.setDate(Number(D));
    d.setHours(Number(H), Number(Mi), 0, 0);
    const ms = d.getTime();
    return Number.isFinite(ms) ? ms : null;
  }
  const tryMs = Date.parse(raw.replace(" ", "T"));
  return Number.isFinite(tryMs) ? tryMs : null;
}

export default function TimeRangePicker({
  value,
  onChange,
  onApply,
  quickRanges,
  placement = "bottom-start",
  size = "sm",
  buttonLabel,
}: TimeRangePickerProps) {
  const now = Date.now();
  const initialFrom = value?.from ?? now - 10 * 60_000;
  const initialTo = value?.to ?? now;

  const [from, setFrom] = useState<number>(initialFrom);
  const [to, setTo] = useState<number>(initialTo);
  const [open, setOpen] = useState(false);

  const [fromText, setFromText] = useState<string>(toEditString(initialFrom));
  const [toText, setToText] = useState<string>(toEditString(initialTo));

  const fromPickerRef = useRef<HTMLInputElement>(null);
  const toPickerRef = useRef<HTMLInputElement>(null);

  const _ranges = quickRanges ?? defaultQuickRanges;

  const displayText = useMemo(() => {
    const fmt = (t: number) => new Date(t).toLocaleString();
    return `${fmt(from)} → ${fmt(to)}`;
  }, [from, to]);

  const apply = () => {
    const next = { from, to };
    onChange?.(next);
    onApply?.(next);
    setOpen(false);
  };

  return (
    <Popover.Root
      open={open}
      onOpenChange={(e) => setOpen(e.open)}
      positioning={{ placement }}
    >
      <Popover.Trigger asChild>
        <Button
          size={size}
          onClick={() => setOpen((s) => !s)}
          variant="outline"
        >
          {buttonLabel ?? displayText}
        </Button>
      </Popover.Trigger>

      <Popover.Positioner>
        <Popover.Content minW="460px">
          <Popover.Arrow />
          <Popover.Body>
            <Flex gap={2} align="stretch">
              {/* left: absolute range */}
              <Box flex="none" w="280px" minW="280px">
                <Text fontSize="sm" mb={2} color="gray.500">
                  绝对时间范围
                </Text>

                <Stack gap={2}>
                  <Box>
                    <Text fontSize="xs" mb={1} color="gray.500">
                      开始
                    </Text>
                    <HStack gap={1}>
                      <Input
                        type="text"
                        value={fromText}
                        onChange={(e) => setFromText(e.target.value)}
                        onBlur={(e) => {
                          const ms = parseEditString(e.target.value);
                          if (ms != null) {
                            setFrom(ms);
                            setFromText(toEditString(ms));
                          }
                        }}
                        placeholder="YYYY-MM-DD HH:mm"
                        size={size}
                        w="206px"
                      />
                      <IconButton
                        aria-label="选择开始时间"
                        variant="outline"
                        size={size}
                        onClick={() => {
                          const el = fromPickerRef.current;
                          if (!el) return;
                          try {
                            (
                              el as HTMLInputElement & {
                                showPicker?: () => void;
                              }
                            ).showPicker?.();
                          } catch {
                            el.focus();
                            el.click();
                          }
                        }}
                      >
                        <Icon icon="mdi:calendar" width={16} height={16} />
                      </IconButton>
                      <input
                        ref={fromPickerRef}
                        type="datetime-local"
                        value={toLocalInputValue(from)}
                        onChange={(e) => {
                          const v = fromLocalInputValue(e.target.value);
                          setFrom(v);
                          setFromText(toEditString(v));
                        }}
                        style={{
                          position: "absolute",
                          opacity: 0,
                          pointerEvents: "none",
                          width: 0,
                          height: 0,
                        }}
                      />
                    </HStack>
                  </Box>

                  <Box>
                    <Text fontSize="xs" mb={1} color="gray.500">
                      结束
                    </Text>
                    <HStack gap={1}>
                      <Input
                        type="text"
                        value={toText}
                        onChange={(e) => setToText(e.target.value)}
                        onBlur={(e) => {
                          const ms = parseEditString(e.target.value);
                          if (ms != null) {
                            setTo(ms);
                            setToText(toEditString(ms));
                          }
                        }}
                        placeholder="YYYY-MM-DD HH:mm"
                        size={size}
                        w="206px"
                      />
                      <IconButton
                        aria-label="选择结束时间"
                        variant="outline"
                        size={size}
                        onClick={() => {
                          const el = toPickerRef.current;
                          if (!el) return;
                          try {
                            (
                              el as HTMLInputElement & {
                                showPicker?: () => void;
                              }
                            ).showPicker?.();
                          } catch {
                            el.focus();
                            el.click();
                          }
                        }}
                      >
                        <Icon icon="mdi:calendar" width={16} height={16} />
                      </IconButton>
                      <input
                        ref={toPickerRef}
                        type="datetime-local"
                        value={toLocalInputValue(to)}
                        onChange={(e) => {
                          const v = fromLocalInputValue(e.target.value);
                          setTo(v);
                          setToText(toEditString(v));
                        }}
                        style={{
                          position: "absolute",
                          opacity: 0,
                          pointerEvents: "none",
                          width: 0,
                          height: 0,
                        }}
                      />
                    </HStack>
                  </Box>

                  <HStack justify="flex-start" pt={1}>
                    <Button
                      size={size}
                      variant="ghost"
                      onClick={() => setOpen(false)}
                    >
                      取消
                    </Button>
                    <Button size={size} colorScheme="blue" onClick={apply}>
                      应用时间范围
                    </Button>
                  </HStack>
                </Stack>
              </Box>

              <Box w="1px" bg="gray.200" _dark={{ bg: "gray.700" }} />

              {/* right: quick ranges */}
              <Box flex="1" minW="180px">
                <Text fontSize="sm" mb={2} color="gray.500">
                  快捷范围
                </Text>
                <Stack gap={1} maxH="280px" overflowY="auto">
                  {(quickRanges ?? defaultQuickRanges).map((r) => (
                    <Button
                      key={`${r.label}-${r.amount}${r.unit}`}
                      variant="ghost"
                      justifyContent="space-between"
                      onClick={() => {
                        const n = Date.now();
                        const f = addRelative(n, r.amount, r.unit);
                        setFrom(f);
                        setTo(n);
                        setFromText(toEditString(f));
                        setToText(toEditString(n));
                        const next = { from: f, to: n };
                        onChange?.(next);
                        onApply?.(next);
                        setOpen(false);
                      }}
                      size={size}
                    >
                      <Text>{r.label}</Text>
                    </Button>
                  ))}
                </Stack>
              </Box>
            </Flex>
          </Popover.Body>
        </Popover.Content>
      </Popover.Positioner>
    </Popover.Root>
  );
}
