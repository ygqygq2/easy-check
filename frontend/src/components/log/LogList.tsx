import { GetLogFiles } from "@bindings/easy-check/internal/services/appservice";
import { types } from "@bindings/easy-check/internal/types";
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

interface LogListProps {
  onClose: () => void;
}

type SortType = "name-asc" | "name-desc" | "time-asc" | "time-desc";

function LogList({ onClose }: LogListProps) {
  const [logFiles, setLogFiles] = useState<types.LogFileInfo[]>([]);
  const [sortedFiles, setSortedFiles] = useState<types.LogFileInfo[]>([]);
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [sortType, setSortType] = useState<SortType>("time-desc");

  // 对文件进行排序
  const sortFiles = (
    files: types.LogFileInfo[],
    sort: SortType
  ): types.LogFileInfo[] => {
    return [...files].sort((a, b) => {
      switch (sort) {
        case "name-asc":
          return a.name.localeCompare(b.name);
        case "name-desc":
          return b.name.localeCompare(a.name);
        case "time-asc":
          return a.modTime - b.modTime;
        case "time-desc":
          return b.modTime - a.modTime;
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
        <Text fontSize="sm" color="app.text.muted">
          共 {logFiles.length} 个日志文件
        </Text>
        <HStack>
          <Text fontSize="sm" color="app.text">
            排序:
          </Text>
          <select
            value={sortType}
            onChange={(e) => setSortType(e.target.value as SortType)}
            style={{
              padding: "0.25rem 0.5rem",
              fontSize: "0.875rem",
              minWidth: "11.25rem",
              borderRadius: "0.375rem",
              borderWidth: "1px",
            }}
          >
            <option value="time-desc">按时间 ↓ (新到旧)</option>
            <option value="time-asc">按时间 ↑ (旧到新)</option>
            <option value="time-desc">按文件名 ↓ (Z到A)</option>
            <option value="name-asc">按文件名 ↑ (A到Z)</option>
          </select>
        </HStack>
      </HStack>

      {/* 文件列表 - 多列网格布局 */}
      <Box flex="1" overflowY="auto" mt="0.75rem">
        {sortedFiles.length === 0 ? (
          <Text
            textAlign="center"
            fontSize="lg"
            color="app.text.muted"
            py="2rem"
          >
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
              const fileTime = new Date(file.modTime * 1000); // 后端传来的是秒，转为毫秒
              const isZeroTime = file.modTime === 0;

              return (
                <GridItem key={file.name}>
                  <Flex
                    direction="column"
                    p="0.75rem"
                    borderWidth="1px"
                    borderRadius="md"
                    borderColor="app.border"
                    boxShadow="sm"
                    bg="app.card"
                    height="4rem"
                    cursor="pointer"
                    transition="all 0.15s ease"
                    _hover={{
                      transform: "translateY(-1px)",
                      boxShadow: "md",
                      bg: "app.surface",
                      borderColor: "blue.400",
                    }}
                    _active={{ transform: "translateY(0)" }}
                    onClick={() => setSelectedFile(file.name)}
                  >
                    <HStack gap="0.5rem" mb="0.25rem">
                      <ChakraIcon color="blue.400" flexShrink={0}>
                        <Icon
                          icon="line-md:document-list"
                          width="14"
                          height="14"
                        />
                      </ChakraIcon>
                      <Text
                        fontWeight="medium"
                        fontSize="0.75rem"
                        color="app.text"
                        overflow="hidden"
                        textOverflow="ellipsis"
                        whiteSpace="nowrap"
                        flex="1"
                        title={file.name}
                      >
                        {file.name}
                      </Text>
                    </HStack>
                    {!isZeroTime && (
                      <Text
                        fontSize="0.625rem"
                        color="app.text.muted"
                        mt="auto"
                      >
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

export default LogList;
