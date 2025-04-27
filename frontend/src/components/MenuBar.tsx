import { Button, Flex, Menu, Portal } from "@chakra-ui/react";
import React from "react";

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
