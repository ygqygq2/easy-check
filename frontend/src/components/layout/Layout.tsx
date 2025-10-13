import {
  CheckForUpdates,
  DisableAutoStart,
  EnableAutoStart,
  IsAutoStartEnabled,
  RestartApp,
} from "@bindings/easy-check/internal/services/appservice";
import { Box, Flex } from "@chakra-ui/react";
import { useEffect, useState } from "react";

import smallLogo from "@/assets/images/logo36x36.png";
import MenuBar from "@/components/MenuBar";
import { ColorModeButton, useColorModeValue } from "@/components/ui/color-mode";

import About from "../About";
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
  const [autoStartEnabled, setAutoStartEnabled] = useState(false);

  // 检查开机自启状态
  useEffect(() => {
    IsAutoStartEnabled()
      .then((enabled) => {
        setAutoStartEnabled(enabled);
      })
      .catch((error) => {
        console.error("检查开机自启状态失败:", error);
      });
  }, []);

  // 切换开机自启状态
  const toggleAutoStart = async () => {
    try {
      if (autoStartEnabled) {
        await DisableAutoStart();
        setAutoStartEnabled(false);
        toaster.create({
          title: "开机自启",
          description: "已禁用开机自启",
          type: "success",
        });
      } else {
        await EnableAutoStart();
        setAutoStartEnabled(true);
        toaster.create({
          title: "开机自启",
          description: "已启用开机自启",
          type: "success",
        });
      }
    } catch (error) {
      toaster.create({
        title: "设置失败",
        description: `${error}`,
        type: "error",
      });
    }
  };

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
          value: "autostart",
          label: autoStartEnabled ? "✓ 开机自启" : "开机自启",
          onClick: toggleAutoStart,
        },
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
            setActiveComponent(
              <About onClose={() => setActiveComponent(children)} />
            );
          },
        },
      ],
    },
  ];

  return (
    <Box height="100vh" display="flex" flexDirection="column" overflow="hidden">
      <Flex
        as="nav"
        bg={navBg}
        color={navColor}
        px={4}
        py={2}
        align="center"
        justify="space-between"
        flexShrink={0}
      >
        <Flex align="center">
          <Box className="small-logo" mr={4}>
            <img src={smallLogo} alt="logo" />
          </Box>
          <MenuBar menus={menus} />
        </Flex>
        <ColorModeButton />
      </Flex>
      <Box flex="1" overflow="hidden" minHeight={0}>
        {activeComponent}
      </Box>
      <Toaster />
    </Box>
  );
}
