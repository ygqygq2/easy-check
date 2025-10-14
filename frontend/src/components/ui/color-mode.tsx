"use client";

import type { IconButtonProps } from "@chakra-ui/react";
import { ClientOnly, IconButton, Skeleton } from "@chakra-ui/react";
import { Icon } from "@iconify/react";
import { useTheme } from "next-themes";
import * as React from "react";

/**
 * Color Mode 管理
 * 使用 next-themes 与 Chakra UI v3 集成
 */

export type ColorMode = "light" | "dark";

export interface UseColorModeReturn {
  colorMode: ColorMode;
  setColorMode: (colorMode: ColorMode) => void;
  toggleColorMode: () => void;
}

/**
 * Color mode hook
 */
export function useColorMode(): UseColorModeReturn {
  const { resolvedTheme, setTheme } = useTheme();

  const toggleColorMode = React.useCallback(() => {
    setTheme(resolvedTheme === "dark" ? "light" : "dark");
  }, [resolvedTheme, setTheme]);

  return {
    colorMode: (resolvedTheme || "light") as ColorMode,
    setColorMode: setTheme as (mode: ColorMode) => void,
    toggleColorMode,
  };
}

/**
 * Color mode value hook - 根据当前模式返回不同的值
 */
export function useColorModeValue<T>(light: T, dark: T): T {
  const { colorMode } = useColorMode();
  return colorMode === "dark" ? dark : light;
}

export function ColorModeIcon() {
  const { colorMode } = useColorMode();
  return colorMode === "dark" ? (
    <Icon icon="mdi:moon-waning-crescent" width="24" height="24" />
  ) : (
    <Icon icon="material-symbols:wb-sunny-outline" width="24" height="24" />
  );
}

type ColorModeButtonProps = Omit<IconButtonProps, "aria-label">;

export const ColorModeButton = React.forwardRef<
  HTMLButtonElement,
  ColorModeButtonProps
>(function ColorModeButton(props, ref) {
  const { toggleColorMode } = useColorMode();
  return (
    <ClientOnly fallback={<Skeleton boxSize="8" />}>
      <IconButton
        onClick={toggleColorMode}
        variant="ghost"
        aria-label="Toggle color mode"
        size="sm"
        ref={ref}
        {...props}
        css={{
          _icon: {
            width: "5",
            height: "5",
          },
        }}
      >
        <ColorModeIcon />
      </IconButton>
    </ClientOnly>
  );
});
