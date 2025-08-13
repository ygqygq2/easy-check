import react from "@ygqygq2/eslint-config/react.mjs";
import tseslint from "typescript-eslint";

export default tseslint.config(...react, {
  ignores: ["*.cjs", "dist", "bindings"],
  rules: {
    // 显式配置有问题的规则
    "@typescript-eslint/no-unused-expressions": [
      "error",
      {
        allowShortCircuit: true,
        allowTernary: true,
        allowTaggedTemplates: true,
      },
    ],
    // 完全禁用 prettier 规则，让 Prettier 独立处理格式化
    "prettier/prettier": "off",
  },
});
