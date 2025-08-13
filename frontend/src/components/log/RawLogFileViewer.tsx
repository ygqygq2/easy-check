import { GetLogFileContent } from "@bindings/easy-check/internal/services/appservice";
import { Box, Text } from "@chakra-ui/react";
import { useEffect, useState } from "react";

import { StatusView } from "../StatusView";
import ActionButton from "../ui/ActionButton";
import { useColorModeValue } from "../ui/color-mode";
import { HeaderWithActions } from "../ui/HeaderWithActions";

function RawLogFileViewer({
  fileName,
  onClose,
}: {
  fileName: string;
  onClose: () => void;
}) {
  const [content, setContent] = useState<string>("");
  const [isLoading, setIsLoading] = useState(true);

  const bgColor = useColorModeValue("gray.50", "gray.800");
  const textColor = useColorModeValue("gray.800", "gray.200");

  useEffect(() => {
    setIsLoading(true);
    GetLogFileContent(fileName, false)
      .then(setContent)
      .finally(() => setIsLoading(false));
  }, [fileName]);

  return (
    <Box
      p="1rem"
      height="100%"
      display="flex"
      flexDirection="column"
      overflow="hidden"
    >
      <HeaderWithActions
        title={`查看日志文件: ${fileName}`}
        actions={<ActionButton label="关闭" onClick={onClose} />}
      />
      {isLoading ? (
        <StatusView message="正在加载文件内容..." isLoading />
      ) : (
        <Box
          flex="1"
          borderWidth="1px"
          borderRadius="md"
          bg={bgColor}
          overflowY="auto"
          mt="0.75rem"
          minHeight={0}
          p="1rem"
        >
          <Text
            fontFamily="monospace"
            whiteSpace="pre-wrap"
            fontSize="0.875rem"
            color={textColor}
            lineHeight="1.4"
          >
            {content}
          </Text>
        </Box>
      )}
    </Box>
  );
}

export default RawLogFileViewer;
