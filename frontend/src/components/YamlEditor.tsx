import {
  GetConfig,
  SaveConfig,
} from "@bindings/easy-check/internal/services/appservice";
import { Box, Flex, Text } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import MonacoEditor from "react-monaco-editor";

import { toaster } from "@/components/ui/toaster";
import { config } from "@/config";
import { loadConfigFromUrl } from "@/lib/load-config";
import { mergeYamlDocuments } from "@/lib/merge-yaml";

import ActionButton from "./ui/ActionButton";
import { useColorMode } from "./ui/color-mode";
import { HeaderWithActions } from "./ui/HeaderWithActions";

interface YamlEditorProps {
  onClose: () => void;
}

const YamlEditor = ({ onClose }: YamlEditorProps) => {
  const [content, setContent] = useState("");
  const { colorMode } = useColorMode();

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      const data = await GetConfig();
      setContent(data);
    } catch (err) {
      toaster.create({
        title: "打开配置失败",
        description: String(err),
        type: "error",
      });
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

  const handleMergeDefault = async () => {
    // 1. 加载默认配置
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    let defaultYaml: any;

    try {
      defaultYaml = await loadConfigFromUrl(config.defaultYamlUrl);
    } catch (err) {
      toaster.create({
        title: "加载默认配置失败",
        description: `无法获取默认配置: ${String(err)}`,
        type: "error",
      });
      return;
    }

    // 2. 验证默认配置
    if (typeof defaultYaml !== "string") {
      toaster.create({
        title: "合并失败",
        description: "默认配置不是有效的字符串",
        type: "error",
      });
      return;
    }

    // 3. 验证当前配置
    if (typeof content !== "string") {
      toaster.create({
        title: "合并失败",
        description: "当前配置不是有效的字符串",
        type: "error",
      });
      return;
    }

    // 4. 合并配置
    let mergedYaml: string;
    try {
      mergedYaml = mergeYamlDocuments(defaultYaml, content);
    } catch (err) {
      toaster.create({
        title: "合并失败",
        description: `合并过程中出现错误: ${String(err)}`,
        type: "error",
      });
      return;
    }

    // 5. 更新内容
    setContent(mergedYaml);
    toaster.create({
      title: "合并成功",
      description: "默认配置已成功合并到当前配置，并保留了注释",
      type: "success",
    });
  };

  const editorOptions = {
    selectOnLineNumbers: true,
    roundedSelection: false,
    readOnly: false,
    cursorStyle: "line" as const,
    automaticLayout: true,
  };

  const editorTheme = colorMode === "dark" ? "vs-dark" : "vs-light";

  return (
    <Box
      height="calc(100vh - 50px)"
      display="flex"
      flexDirection="column"
      p={4}
    >
      <HeaderWithActions
        title="配置编辑器"
        actions={
          <>
            <ActionButton label="取消" onClick={onClose} />
            <ActionButton label="合并默认配置" onClick={handleMergeDefault} />
            <ActionButton label="重新加载" onClick={loadConfig} />
            <ActionButton label="保存配置" onClick={handleSave} />
          </>
        }
      />

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
