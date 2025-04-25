import tseslint from 'typescript-eslint';

import react from '@ygqygq2/eslint-config/react.mjs';

export default tseslint.config(...react, {
  ignores: ['dist', 'wailsjs'],
  rules: {
    // 显式配置有问题的规则
    '@typescript-eslint/no-unused-expressions': [
      'error',
      {
        allowShortCircuit: true,
        allowTernary: true,
        allowTaggedTemplates: true
      }
    ]
  },
});
