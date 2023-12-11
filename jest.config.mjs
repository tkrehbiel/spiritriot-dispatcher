export default {
  verbose: true,
  testMatch: ['**/__tests__/**/*.mjs?(x)'],
  testEnvironment: 'node',
  transform: {
    '^.+\\.mjs$': 'babel-jest',
  },
};
