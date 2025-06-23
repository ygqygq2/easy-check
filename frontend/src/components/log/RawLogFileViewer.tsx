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
    <Box p={4}>
      <HeaderWithActions
        title={`查看日志文件: ${fileName}`}
        actions={<ActionButton label="关闭" onClick={onClose} />}
      />
      {isLoading ? (
        <StatusView message="正在加载文件内容..." isLoading />
      ) : (
        <Box
          borderWidth="1px"
          borderRadius="md"
          p={4}
          bg={bgColor}
          maxH="80vh"
          overflow="auto"
        >
          <Text
            fontFamily="monospace"
            whiteSpace="pre-wrap"
            fontSize="sm"
            color={textColor}
          >
            {content}
          </Text>
        </Box>
      )}
    </Box>
  );
}

export default RawLogFileViewer;
