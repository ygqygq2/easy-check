import { Box, Button, Flex, Spinner, Text } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import MonacoEditor from "react-monaco-editor";

import { toaster } from "@/components/ui/toaster";

import { GetConfig, SaveConfig } from "../../wailsjs/go/main/App";
import { useColorMode, useColorModeValue } from "./ui/color-mode";

interface YamlEditorProps {
  onClose: () => void;
}

const YamlEditor = ({ onClose }: YamlEditorProps) => {
  const buttonBg = useColorModeValue("gray.200", "gray.700"); // 按钮背景色
  const buttonColor = useColorModeValue("gray.800", "white");
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(true);
  const { colorMode } = useColorMode();

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      setLoading(true);
      const data = await GetConfig();
      setContent(data);
    } catch (err) {
      toaster.create({
        title: "打开配置失败",
        description: String(err),
        type: "error",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      await SaveConfig(content);
      toaster.create({
        title: "保存成功",
        description: "配置已成功保存",
        type: "success",
      });
      onClose();
    } catch (err) {
      toaster.create({
        title: "保存失败",
        description: String(err),
        type: "error",
      });
    }
  };

  const editorOptions = {
    selectOnLineNumbers: true,
    roundedSelection: false,
    readOnly: false,
    cursorStyle: "line" as const,
    automaticLayout: true,
  };

  const editorTheme = colorMode === "dark" ? "vs-dark" : "vs-light";

  if (loading) {
    return (
      <Flex justify="center" align="center" height="100%">
        <Spinner size="xl" />
      </Flex>
    );
  }

  return (
    <Box height="calc(100vh - 50px)" display="flex" flexDirection="column" p={4}>
      <Flex justify="space-between" mb={4}>
        <Text fontSize="xl">配置编辑器</Text>
        <Flex gap={4}>
          <Button
            bg={buttonBg}
            color={buttonColor}
            _hover={{ bg: useColorModeValue("gray.300", "gray.600") }}
            onClick={loadConfig}
          >
            重新加载
          </Button>
          <Button
            bg={buttonBg}
            color={buttonColor}
            _hover={{ bg: useColorModeValue("gray.300", "gray.600") }}
            onClick={handleSave}
          >
            保存配置
          </Button>
        </Flex>
      </Flex>

      <Box flex="1" borderWidth="1px" borderRadius="md" overflow="hidden">
        <MonacoEditor
          height="100%"
          language="yaml"
          theme={editorTheme}
          value={content}
          options={editorOptions}
          onChange={setContent}
        />
      </Box>
    </Box>
  );
};

export default YamlEditor;
