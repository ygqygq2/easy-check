import { HStack, Slider, Switch } from "@chakra-ui/react";

import ActionButton from "../ui/ActionButton";

interface LogControlPanelProps {
  isRealtime: boolean;
  onRealtimeChange: (value: boolean) => void;
  isLatest: boolean;
  onClose: () => void;
  updateInterval: number;
  onUpdateIntervalChange: (value: number) => void;
}

export const LogControlPanel = ({
  isRealtime,
  onRealtimeChange,
  isLatest,
  onClose,
  updateInterval,
  onUpdateIntervalChange,
}: LogControlPanelProps) => {
  return (
    <HStack gap={4}>
      {isLatest && (
        <HStack gap={4}>
          {isRealtime && (
            <Slider.Root
              min={10}
              max={60}
              step={10}
              value={[updateInterval]}
              onValueChange={(e) => onUpdateIntervalChange(e.value[0])}
              size="sm"
              w="200px"
            >
              <HStack>
                <Slider.Label>更新频率</Slider.Label>
                <Slider.ValueText />
                <Slider.Label>s</Slider.Label>
              </HStack>
              <Slider.Control>
                <Slider.Track>
                  <Slider.Range />
                </Slider.Track>
                <Slider.Thumbs />
              </Slider.Control>
            </Slider.Root>
          )}

          <Switch.Root
            checked={isRealtime}
            onCheckedChange={(e) => onRealtimeChange(e.checked)}
          >
            <Switch.HiddenInput />
            <Switch.Control />
            <Switch.Label>实时更新</Switch.Label>
          </Switch.Root>
        </HStack>
      )}
      <ActionButton label="关闭" onClick={onClose} />
    </HStack>
  );
};
