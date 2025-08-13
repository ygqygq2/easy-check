import { GetLogFiles } from "@bindings/easy-check/internal/services/appservice";
import {
  Box,
  Flex,
  Text,
  HStack,
  Icon as ChakraIcon,
  Grid,
  GridItem,
} from "@chakra-ui/react";
import { Icon } from "@iconify/react";
import { useEffect, useState, ChangeEvent } from "react";

import ActionButton from "../ui/ActionButton";
import { HeaderWithActions } from "../ui/HeaderWithActions";
import { toaster } from "../ui/toaster";
import RawLogFileViewer from "./RawLogFileViewer";

interface LogListComponentProps {
  onClose: () => void;
}

type SortType = "name-asc" | "name-desc" | "time-asc" | "time-desc";

function LogListComponent({ onClose }: LogListComponentProps) {
  const [logFiles, setLogFiles] = useState<string[]>([]);
  const [sortedFiles, setSortedFiles] = useState<string[]>([]);
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [sortType, setSortType] = useState<SortType>("time-desc");

  // 从文件名提取时间戳
  const extractTimeFromFileName = (fileName: string): Date | null => {
    // 匹配 check-log-2025-08-11T15-51-47.344.txt 格式
    if (fileName.startsWith("check-log-") && fileName.endsWith(".txt")) {
      const timeStr = fileName
        .replace("check-log-", "")
        .replace(".txt", "")
        .replace("T", " ")
        .replace(/-/g, ":");

      const parsedTime = new Date(timeStr);
      return isNaN(parsedTime.getTime()) ? null : parsedTime;
    }
    return null;
  };

  // 对文件进行排序
  const sortFiles = (files: string[], sort: SortType): string[] => {
    return [...files].sort((a, b) => {
      switch (sort) {
        case "name-asc":
          return a.localeCompare(b);
        case "name-desc":
          return b.localeCompare(a);
        case "time-asc": {
          const timeA = extractTimeFromFileName(a);
          const timeB = extractTimeFromFileName(b);
          if (timeA && timeB) {
            return timeA.getTime() - timeB.getTime();
          }
          return a.localeCompare(b);
        }
        case "time-desc": {
          const timeA = extractTimeFromFileName(a);
          const timeB = extractTimeFromFileName(b);
          if (timeA && timeB) {
            return timeB.getTime() - timeA.getTime();
          }
          return b.localeCompare(a);
        }
        default:
          return 0;
      }
    });
  };

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

  useEffect(() => {
    setSortedFiles(sortFiles(logFiles, sortType));
  }, [logFiles, sortType]);

  const handleSortChange = (e: ChangeEvent<HTMLSelectElement>) => {
    setSortType(e.target.value as SortType);
  };

  if (selectedFile) {
    return (
      <RawLogFileViewer
        fileName={selectedFile}
        onClose={() => setSelectedFile(null)}
      />
    );
  }

  return (
    <Box
      p="1rem"
      height="100%"
      display="flex"
      flexDirection="column"
      overflow="hidden"
    >
      <HeaderWithActions
        title="日志文件列表"
        actions={<ActionButton label="关闭" onClick={onClose} />}
      />

      {/* 排序控件 */}
      <HStack mb="0.75rem" justify="space-between" flexShrink={0}>
        <Text fontSize="sm" color="gray.600">
          共 {logFiles.length} 个日志文件
        </Text>
        <HStack>
          <Text fontSize="sm">排序:</Text>
          <Box>
            <select
              style={{
                padding: "0.25rem 0.5rem",
                borderRadius: "0.25rem",
                border: "1px solid #d1d5db",
                fontSize: "0.875rem",
                minWidth: "11.25rem",
              }}
              value={sortType}
              onChange={handleSortChange}
            >
              <option value="time-desc">按时间 ↓ (新到旧)</option>
              <option value="time-asc">按时间 ↑ (旧到新)</option>
              <option value="name-desc">按文件名 ↓ (Z到A)</option>
              <option value="name-asc">按文件名 ↑ (A到Z)</option>
            </select>
          </Box>
        </HStack>
      </HStack>

      {/* 文件列表 - 多列网格布局 */}
      <Box flex="1" overflowY="auto" mt="0.75rem">
        {sortedFiles.length === 0 ? (
          <Text textAlign="center" fontSize="lg" color="gray.500" py="2rem">
            没有日志文件
          </Text>
        ) : (
          <Grid
            templateColumns={{
              base: "1fr",
              md: "repeat(2, 1fr)",
              lg: "repeat(3, 1fr)",
              xl: "repeat(4, 1fr)",
              "2xl": "repeat(5, 1fr)",
            }}
            gap="0.75rem"
            pb="1.5rem"
          >
            {sortedFiles.map((file) => {
              const fileTime = extractTimeFromFileName(file);
              return (
                <GridItem key={file}>
                  <Flex
                    direction="column"
                    p="0.75rem"
                    borderWidth="1px"
                    borderRadius="md"
                    boxShadow="sm"
                    bg="white"
                    height="4rem"
                    cursor="pointer"
                    transition="all 0.15s ease"
                    _hover={{
                      transform: "translateY(-1px)",
                      boxShadow: "md",
                      bg: "gray.50",
                      borderColor: "blue.200",
                    }}
                    _active={{ transform: "translateY(0)" }}
                    onClick={() => setSelectedFile(file)}
                  >
                    <HStack gap="0.5rem" mb="0.25rem">
                      <ChakraIcon color="blue.500" flexShrink={0}>
                        <Icon
                          icon="line-md:document-list"
                          width="14"
                          height="14"
                        />
                      </ChakraIcon>
                      <Text
                        fontWeight="medium"
                        fontSize="0.75rem"
                        overflow="hidden"
                        textOverflow="ellipsis"
                        whiteSpace="nowrap"
                        flex="1"
                        title={file}
                      >
                        {file}
                      </Text>
                    </HStack>
                    {fileTime && (
                      <Text fontSize="0.625rem" color="gray.500" mt="auto">
                        {fileTime.toLocaleString("zh-CN", {
                          month: "2-digit",
                          day: "2-digit",
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                      </Text>
                    )}
                  </Flex>
                </GridItem>
              );
            })}
          </Grid>
        )}
      </Box>
    </Box>
  );
}

export default LogListComponent;
