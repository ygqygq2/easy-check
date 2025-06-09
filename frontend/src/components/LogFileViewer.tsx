import { GetLogFileContent } from "@bindings/easy-check/internal/services/appservice";
import { Box, Spinner, Switch, Text, VStack } from "@chakra-ui/react";
import { useEffect, useRef, useState } from "react";

import ActionButton from "./ui/ActionButton";
import { HeaderWithActions } from "./ui/HeaderWithActions";
import { toaster } from "./ui/toaster";

interface LogFileViewerProps {
  fileName: string;
  onClose: () => void;
  isLatest?: boolean; // 是否是最新日志
}

function LogFileViewer({
  fileName,
  onClose,
  isLatest = false,
}: LogFileViewerProps) {
  const [content, setContent] = useState<string>("");
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRealtime, setIsRealtime] = useState<boolean>(true); // 是否实时滚动
  const contentRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const fetchFileContent = async () => {
      try {
        setIsLoading(true);
        const content = await GetLogFileContent(fileName, isLatest);
        setContent(content);
      } catch (err) {
        toaster.create({
          title: "加载文件失败",
          description: `无法加载文件内容: ${err}`,
          type: "error",
        });
      } finally {
        setIsLoading(false);
      }
    };

    fetchFileContent();

    if (isLatest && isRealtime) {
      const intervalId = setInterval(fetchFileContent, 2000); // 每2秒刷新内容
      return () => clearInterval(intervalId); // 清除定时器
    }
  }, [fileName, isLatest, isRealtime]);

  useEffect(() => {
    if (!isLoading && isLatest && isRealtime && contentRef.current) {
      // 滚动到最底部
      contentRef.current.scrollTop = contentRef.current.scrollHeight;
    }
  }, [isLoading, isLatest, isRealtime, content]);

  return (
    <Box p={4}>
      <HeaderWithActions
        title={`查看日志文件: ${fileName}`}
        actions={
          <>
            {isLatest && (
              <Switch.Root
                checked={isRealtime}
                onCheckedChange={(e) => setIsRealtime(e.checked)}
                mr={4}
              >
                <Switch.HiddenInput />
                <Switch.Control />
                <Switch.Label>实时滚动</Switch.Label>
              </Switch.Root>
            )}
            <ActionButton label="关闭" onClick={onClose} />
          </>
        }
      />
      <VStack align="stretch" gap={2}>
        {isLoading ? (
          <Box textAlign="center" py={4}>
            <Spinner size="lg" />
            <Text mt={2} color="gray.500">
              正在加载文件内容...
            </Text>
          </Box>
        ) : content ? (
          <Box
            ref={contentRef}
            maxHeight={600} // 限制最大高度
            overflowY="auto" // 添加滚动条
            borderWidth="1px"
            borderRadius="md"
            p={4}
            bg="gray.50"
          >
            <Text whiteSpace="pre-wrap">{content}</Text>
          </Box>
        ) : (
          <Text color="gray.500" textAlign="center">
            文件内容为空
          </Text>
        )}
      </VStack>
    </Box>
  );
}

export default LogFileViewer;
