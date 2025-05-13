import { Box, Flex } from "@chakra-ui/react";

import smallLogo from "@/assets/images/logo36x36.png";
import MenuBar from "@/components/MenuBar";
import { ColorModeButton, useColorModeValue } from "@/components/ui/color-mode";

import { CheckForUpdates, RestartApp } from "../../../wailsjs/go/main/App";

interface LayoutProps {
  children: React.ReactNode;
}

export function Layout({ children }: LayoutProps) {
  const navBg = useColorModeValue("gray.200", "gray.700");
  const navColor = useColorModeValue("gray.800", "gray.100");
  const menus = [
    {
      label: "文件",
      items: [
        { value: "open", label: "打开配置", onClick: () => alert("打开配置") },
        { value: "exit", label: "退出", onClick: () => alert("退出") },
      ],
    },
    {
      label: "查看",
      items: [
        {
          value: "log-list",
          label: "日志列表",
          onClick: () => alert("日志列表"),
        },
        {
          value: "latest-log",
          label: "最新日志",
          onClick: () => alert("最新日志"),
        },
      ],
    },
    {
      label: "帮助",
      items: [
        {
          value: "update",
          label: "检查更新",
          onClick: () => {
            CheckForUpdates().then((res) => {
              alert(res);
            });
          },
        },
        {
          value: "restart",
          label: "重启",
          onClick: () => {
            RestartApp().then((res) => {
              alert(res);
            });
          },
        },
        { value: "about", label: "关于", onClick: () => alert("关于") },
      ],
    },
  ];

  return (
    <Box>
      <Flex as="nav" bg={navBg} color={navColor} px={4} py={2} align="center" justify="space-between">
        <Flex align="center">
          <Box className="small-logo" mr={4}>
            <img src={smallLogo} alt="logo" />
          </Box>
          <MenuBar menus={menus} />
        </Flex>
        <ColorModeButton />
      </Flex>
      {children}
    </Box>
  );
}
