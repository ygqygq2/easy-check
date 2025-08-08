import {
  CheckForUpdates,
  RestartApp,
} from "@bindings/easy-check/internal/services/appservice";
import { Box, Flex } from "@chakra-ui/react";
import { useState } from "react";

import smallLogo from "@/assets/images/logo36x36.png";
import MenuBar from "@/components/MenuBar";
import { ColorModeButton, useColorModeValue } from "@/components/ui/color-mode";

import LogFileViewer from "../log/LogFileViewer";
import LogListComponent from "../log/LogListComponent";
import { Toaster, toaster } from "../ui/toaster";
import YamlEditor from "../YamlEditor";

interface LayoutProps {
  children: React.ReactNode;
}

export function Layout({ children }: LayoutProps) {
  const navBg = useColorModeValue("gray.200", "gray.700");
  const navColor = useColorModeValue("gray.800", "gray.100");

  const [activeComponent, setActiveComponent] =
    useState<React.ReactNode>(children);

  const menus = [
    {
      label: "文件",
      items: [
        {
          value: "open",
          label: "打开配置",
          onClick: () =>
            setActiveComponent(
              <YamlEditor onClose={() => setActiveComponent(children)} />
            ),
        },
        { value: "exit", label: "退出", onClick: () => alert("退出") },
      ],
    },
    {
      label: "查看",
      items: [
        {
          value: "log-list",
          label: "日志列表",
          onClick: () =>
            setActiveComponent(
              <LogListComponent onClose={() => setActiveComponent(children)} />
            ),
        },
        {
          value: "latest-log",
          label: "最新日志",
          onClick: () => {
            setActiveComponent(
              <LogFileViewer
                fileName=""
                onClose={() => setActiveComponent(children)}
                isLatest={true}
              />
            );
          },
        },
      ],
    },
    {
      label: "帮助",
      items: [
        {
          value: "update",
          label: "检查更新",
          onClick: async () => {
            try {
              const res = await CheckForUpdates();
              toaster.create({
                title: "检查更新",
                description: res,
                type: "info", // 根据需要可以改为 "success" 或 "warning"
              });
            } catch (error) {
              toaster.create({
                title: "检查更新失败",
                description: `发生错误: ${error}`,
                type: "error",
              });
            }
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
        {
          value: "about",
          label: "关于",
          onClick: () => {
            toaster.create({
              title: "Toaster 测试",
              description: "这是一个测试通知，Toaster 正常工作！",
              type: "info",
            });
          },
        },
      ],
    },
  ];

  return (
    <Box minH="100vh" display="flex" flexDirection="column" overflow="hidden">
      <Flex
        as="nav"
        bg={navBg}
        color={navColor}
        px={4}
        py={2}
        align="center"
        justify="space-between"
      >
        <Flex align="center">
          <Box className="small-logo" mr={4}>
            <img src={smallLogo} alt="logo" />
          </Box>
          <MenuBar menus={menus} />
        </Flex>
        <ColorModeButton />
      </Flex>
      <Box flex="1" overflow="hidden">
        {activeComponent}
      </Box>
      <Toaster />
    </Box>
  );
}
