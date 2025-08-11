import {
  Box,
  HStack,
  Text,
  Image,
  Button,
  Flex,
  Stack,
} from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { useConfig } from "@/hooks/useConfig";
import { CheckForUpdates } from "@bindings/easy-check/internal/services/appservice";
import smallLogo from "@/assets/images/logo36x36.png";

interface AboutProps {
  onClose: () => void;
}

const About = ({ onClose }: AboutProps) => {
  const { appInfo, loadAppInfo } = useConfig();
  const [copied, setCopied] = useState(false);

  // 只在appInfo为空时加载一次，避免无限循环
  useEffect(() => {
    if (!appInfo) {
      loadAppInfo();
    }
  }, []); // 空依赖数组

  const handleOpenRepository = () => {
    if (appInfo?.repository) {
      window.open(appInfo.repository, "_blank");
    }
  };

  // 复制版本信息到剪贴板（VSCode风格）
  const handleCopyInfo = async () => {
    if (!appInfo) return;

    const infoText = [
      `${appInfo.appName}`,
      `版本: ${appInfo.appVersion}`,
      `作者: ${appInfo.author}`,
      `构建时间: ${appInfo.buildTime}`,
      `平台: ${appInfo.platformInfo.os}/${appInfo.platformInfo.arch}`,
      `代码仓库: ${appInfo.repository}`,
    ].join("\n");

    try {
      await navigator.clipboard.writeText(infoText);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("复制失败:", err);
    }
  };

  if (!appInfo) {
    return (
      <Box
        p={4}
        w="400px"
        h="300px"
        display="flex"
        alignItems="center"
        justifyContent="center"
      >
        <Text>加载中...</Text>
      </Box>
    );
  }

  return (
    <Box p={4} w="400px" h="300px" display="flex" flexDirection="column">
      {/* 顶部：图标和标题 */}
      <HStack gap={3} mb={4}>
        <Image src={smallLogo} alt="Logo" w={12} h={12} />
        <Box>
          <Text fontSize="lg" fontWeight="bold">
            {appInfo.appName}
          </Text>
          <Text fontSize="sm" color="gray.600">
            {appInfo.description}
          </Text>
        </Box>
      </HStack>

      {/* 中间：紧凑的信息列表 */}
      <Box flex={1} fontSize="sm" lineHeight="1.6">
        <Text>版本: {appInfo.appVersion}</Text>
        <Text>构建时间: {appInfo.buildTime}</Text>
        <Text>作者: {appInfo.author}</Text>
        <Text>
          平台: {appInfo.platformInfo.os}/{appInfo.platformInfo.arch}
        </Text>
        <Text>
          代码仓库:{" "}
          <Button
            variant="ghost"
            size="sm"
            onClick={handleOpenRepository}
            color="blue.500"
            p={0}
            h="auto"
            fontSize="sm"
            textDecoration="underline"
          >
            GitHub
          </Button>
        </Text>
        <Text mt={2} fontSize="xs" color="gray.500">
          {appInfo.copyright}
        </Text>
        <Text fontSize="xs" color="gray.500">
          {appInfo.license}
        </Text>
      </Box>

      {/* 底部：VSCode风格的按钮 */}
      <Flex
        justify="space-between"
        mt={4}
        pt={3}
        borderTop="1px solid"
        borderColor="gray.200"
      >
        <Button
          variant="ghost"
          size="sm"
          onClick={handleCopyInfo}
          colorScheme={copied ? "green" : "gray"}
        >
          {copied ? "已复制" : "复制"}
        </Button>
        <Button onClick={onClose} colorScheme="blue" size="sm">
          确定
        </Button>
      </Flex>
    </Box>
  );
};

export default About;
