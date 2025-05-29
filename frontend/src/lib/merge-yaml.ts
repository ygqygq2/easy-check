import * as yaml from "yaml";
import { Document, isMap, isPair, Pair, YAMLMap } from "yaml";

// 扩展 Pair 类型以支持注释
interface CommentedPair<K = unknown, V = unknown> extends Pair<K, V> {
  comment?: string | null;
  commentBefore?: string | null;
}

// 类型守卫：判断 Pair 是否为 CommentedPair
function isCommentedPair<K = unknown, V = unknown>(
  pair: Pair<K, V>
): pair is CommentedPair<K, V> {
  return true; // 运行时属性总会有，类型断言用
}

/**
 * 合并两个 YAML 文档，保留注释
 * @param defaultYaml 默认配置的 YAML 字符串
 * @param currentYaml 当前配置的 YAML 字符串
 * @returns 合并后的 YAML 字符串
 */
export const mergeYamlDocuments = (
  defaultYaml: string,
  currentYaml: string
): string => {
  const defaultDoc = yaml.parseDocument(defaultYaml);
  const currentDoc = yaml.parseDocument(currentYaml);

  // 克隆默认配置文档
  const mergedDoc = defaultDoc.clone();

  // 确保 mergedDoc.contents 是 YAMLMap
  ensureYAMLMap(mergedDoc);

  // 合并内容
  if (isMap(currentDoc.contents)) {
    mergeYAMLMaps(
      mergedDoc.contents as YAMLMap,
      currentDoc.contents as YAMLMap
    );
  }

  // 合并文件级注释
  mergedDoc.commentBefore =
    defaultDoc.commentBefore || currentDoc.commentBefore || null;

  return mergedDoc.toString();
};

/**
 * 确保文档内容是 YAMLMap 类型
 */
function ensureYAMLMap(doc: Document): void {
  if (!(doc.contents instanceof YAMLMap)) {
    doc.contents = doc.createNode({}) as YAMLMap;
  }
}

/**
 * 合并两个 YAMLMap，保留注释
 */
function mergeYAMLMaps(targetMap: YAMLMap, sourceMap: YAMLMap): void {
  for (const item of sourceMap.items) {
    if (!isPair(item)) continue;
    const key = String(item.key);

    // 查找目标中的同名 Pair
    const targetPair = targetMap.items.find(
      (p): p is CommentedPair => isPair(p) && String(p.key) === key
    );

    if (
      targetPair &&
      item.value instanceof YAMLMap &&
      targetPair.value instanceof YAMLMap
    ) {
      // 递归合并嵌套对象
      mergeYAMLMaps(targetPair.value, item.value);
    } else if (targetPair) {
      // 类型断言为 CommentedPair
      if (isCommentedPair(targetPair) && isCommentedPair(item)) {
        targetPair.value = item.value;
        targetPair.comment = item.comment ?? targetPair.comment;
        targetPair.commentBefore =
          item.commentBefore ?? targetPair.commentBefore;
      } else {
        targetPair.value = item.value;
      }
    } else {
      // 新增项，直接添加
      targetMap.add(item);
    }
  }
}
