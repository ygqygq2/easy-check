import { GetLogFiles } from "@bindings/easy-check/internal/services/appservice";
import { Box, Flex, Text, VStack } from "@chakra-ui/react";
import { Icon } from "@iconify/react";
import { useEffect, useState } from "react";

import ActionButton from "../ui/ActionButton";
import { HeaderWithActions } from "../ui/HeaderWithActions";
import { toaster } from "../ui/toaster";
import RawLogFileViewer from "./RawLogFileViewer";

interface LogListComponentProps {
  onClose: () => void;
}

function LogListComponent({ onClose }: LogListComponentProps) {
  const [logFiles, setLogFiles] = useState<string[]>([]);
  const [selectedFile, setSelectedFile] = useState<string | null>(null);

  useEffect(() => {
    const fetchLogFiles = async () => {
      try {
        const files = await GetLogFiles();
        setLogFiles(files);
      } catch (err) {
        toaster.create({
          title: "获取日志文件失败",
          description: `无法加载日志文件列表, ${err}`,
          type: "error",
        });
      }
    };

    fetchLogFiles();
  }, []);

  if (selectedFile) {
    return (
      <RawLogFileViewer
        fileName={selectedFile}
        onClose={() => setSelectedFile(null)}
      />
    );
  }

  return (
    <Box p={4}>
      <HeaderWithActions
        title="日志文件列表"
        actions={<ActionButton label="关闭" onClick={onClose} />}
      />
      <VStack gap={1} align="stretch">
        {logFiles.length === 0 ? (
          <Text textAlign="center" fontSize="lg">
            没有日志文件
          </Text>
        ) : (
          logFiles
            .reduce((rows, file, index) => {
              if (index % 2 === 0) rows.push([]);
              rows[rows.length - 1].push(file);
              return rows;
            }, [] as string[][])
            .map((row, rowIndex) => (
              <Flex key={rowIndex} gap={4}>
                {row.map((file, colIndex) => (
                  <Flex
                    key={colIndex}
                    flex="1"
                    align="center"
                    p={4}
                    borderWidth="1px"
                    borderRadius="md"
                    boxShadow="sm"
                    _hover={{ transform: "scale(1.02)" }}
                    transition="all 0.2s"
                    onClick={() => setSelectedFile(file)}
                  >
                    <Icon icon="line-md:file-document" width="24" height="24" />
                    <Text flex="1" fontWeight="medium">
                      {file}
                    </Text>
                  </Flex>
                ))}
              </Flex>
            ))
        )}
      </VStack>
    </Box>
  );
}

export default LogListComponent;
