import nextCoreWebVitals from 'eslint-config-next/core-web-vitals';

export default [
  ...nextCoreWebVitals,
  {
    ignores: ['.next/**', 'node_modules/**', 'next-env.d.ts', 'out/**', 'build/**'],
  },
  {
    rules: {
      'react-hooks/set-state-in-effect': 'warn',
      'react-hooks/purity': 'warn',
      '@next/next/no-img-element': 'warn',
      'import/no-anonymous-default-export': 'off',
    },
  },
];
