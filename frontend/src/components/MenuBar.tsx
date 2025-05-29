import { Button, Flex, Menu, Portal } from "@chakra-ui/react";
import React from "react";

import { useColorModeValue } from "./ui/color-mode";

interface MenuItemProps {
  value: string;
  label: string;
  onClick: () => void;
}

interface MenuBarProps {
  menus: {
    label: string;
    items: MenuItemProps[];
  }[];
}

const MenuBar: React.FC<MenuBarProps> = ({ menus }) => {
  const activeBg = useColorModeValue("gray.100", "gray.700");

  return (
    <Flex gap={2} w="100%">
      {menus.map((menu, index) => (
        <Menu.Root key={index}>
          <Menu.Trigger asChild>
            <Button
              variant="ghost"
              fontSize="sm"
              fontWeight="normal"
              py={2}
              px={2}
              mr={2}
              _active={{ bg: "gray.300", transform: "scale(0.98)" }}
            >
              {menu.label}
            </Button>
          </Menu.Trigger>
          <Portal>
            <Menu.Positioner>
              <Menu.Content>
                {menu.items.map((item) => (
                  <Menu.Item
                    key={item.value}
                    value={item.value}
                    onClick={item.onClick}
                    _active={{ bg: activeBg }}
                  >
                    {item.label}
                  </Menu.Item>
                ))}
              </Menu.Content>
            </Menu.Positioner>
          </Portal>
        </Menu.Root>
      ))}
    </Flex>
  );
};

export default MenuBar;
