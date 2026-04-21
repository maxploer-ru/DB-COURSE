# React + TypeScript + Vite

Currently, two official plugins are available:

## React Compiler
## Что уже есть

- React + TypeScript + Vite
- Роутинг (`/`, `/videos`, `/videos/:videoId`, `/login`, `/register`)
- Базовый auth-store на `zustand`
- API-клиент на `axios`
- Загрузка списка видео и страницы видео с комментариями
## Expanding the ESLint configuration

## Структура

- `src/app` - bootstrap, роутинг, провайдеры
- `src/features` - функциональные модули (`auth`)
- `src/entities` - базовые сущности (`video`)
- `src/pages` - страницы
```js
export default defineConfig([
      // Other configs...
    ],
    languageOptions: {
      parserOptions: {
```js
// eslint.config.js
import reactX from 'eslint-plugin-react-x'
import reactDom from 'eslint-plugin-react-dom'
export default defineConfig([
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...
      // Enable lint rules for React
      reactX.configs['recommended-typescript'],
      // Enable lint rules for React DOM
      reactDom.configs.recommended,
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
      },
      // other options...
    },
  },
])
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...

## Быстрый старт

```bash
npm install
npm run dev
```

## Проверки

```bash
npm run typecheck
npm run lint
npm run test
npm run build
```

## Настройки окружения

Скопируйте `.env.example` в `.env` и при необходимости поменяйте URL API.

```bash
cp .env.example .env
```

По умолчанию используется `VITE_API_BASE_URL=/api` (через proxy Vite на `http://localhost:8080`).

