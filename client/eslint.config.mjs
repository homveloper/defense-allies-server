import { dirname } from "path";
import { fileURLToPath } from "url";
import { FlatCompat } from "@eslint/eslintrc";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const eslintConfig = [
  ...compat.extends("next/core-web-vitals", "next/typescript"),
  {
    files: [
      "src/game/**/*.{ts,tsx}",
      "src/components/game/**/*.{ts,tsx}",
      "src/components/games/**/*.{ts,tsx}",
      "src/components/minimal-legion/**/*.{ts,tsx}",
      "src/components/ability-arena/**/*.{ts,tsx}",
      "src/app/minimal-legion/**/*.{ts,tsx}",
      "src/app/ability-arena/**/*.{ts,tsx}"
    ],
    rules: {
      "@typescript-eslint/no-explicit-any": "off",
      "@typescript-eslint/no-unused-vars": "off"
    }
  }
];

export default eslintConfig;
