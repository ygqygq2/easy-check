import react from "@ygqygq2/eslint-config/react.mjs";
import tseslint from "typescript-eslint";

export default tseslint.config(...react, {
  ignores: ["*.cjs", "dist", "wailsjs"],
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
    "prettier/prettier": [
      "warn",
      {
        trailingComma: "es5",
        singleAttributePerLine: false,
      },
    ],
  },
});
