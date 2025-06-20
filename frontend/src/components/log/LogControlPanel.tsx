import {
  HStack,
  Slider,
  SliderThumb,
  SliderTrack,
  Switch,
  Text,
} from "@chakra-ui/react";
import { useState } from "react";

import ActionButton from "../ui/ActionButton";

interface LogControlPanelProps {
  isRealtime: boolean;
  onRealtimeChange: (value: boolean) => void;
  isLatest: boolean;
  onClose: () => void;
}

export const LogControlPanel = ({
  isRealtime,
  onRealtimeChange,
  isLatest,
  onClose,
}: LogControlPanelProps) => {
  const [updateInterval, setUpdateInterval] = useState([10]);

  return (
    <HStack gap={4}>
      {isLatest && (
        <HStack>
          {isRealtime && (
            <Slider.Root
              min={10}
              max={60}
              step={10}
              value={updateInterval}
              onValueChange={(e) => setUpdateInterval(e.value)}
              size="sm"
            >
              <HStack justify="space-between">
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
