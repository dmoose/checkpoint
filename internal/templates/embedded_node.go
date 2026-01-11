package templates

// Node.js API template
const nodeApiProjectYml = `schema_version: "1"
name: "{{project_name}}"
type: node-api
purpose: |
  Node.js API server.
  (Update this with your project's specific purpose)

architecture:
  overview: |
    Express/Fastify API with routes in src/routes/.
    Business logic in src/services/, data access in src/models/.

  key_paths:
    entry_point: src/index.js
    routes: src/routes/
    services: src/services/
    models: src/models/
    config: src/config/

languages:
  primary: javascript
  version: "node 18+"

dependencies:
  external: []
`

const nodeApiToolsYml = `schema_version: "1"

build:
  default:
    command: npm run build
    notes: Build for production (if using TypeScript)

test:
  default:
    command: npm test
    notes: Run all tests

  watch:
    command: npm run test:watch
    notes: Run tests in watch mode

  coverage:
    command: npm run test:coverage
    notes: Run tests with coverage

lint:
  default:
    command: npm run lint
    notes: Run ESLint

  fix:
    command: npm run lint:fix
    notes: Auto-fix linting issues

check:
  default:
    command: npm run lint && npm test
    notes: Lint and test

run:
  default:
    command: npm start
    notes: Start the server

  dev:
    command: npm run dev
    notes: Start in development mode with hot reload

maintenance:
  deps:
    command: npm install
    notes: Install dependencies

  update:
    command: npm update
    notes: Update dependencies

  audit:
    command: npm audit
    notes: Check for vulnerabilities
`

const nodeApiGuidelinesYml = `schema_version: "1"

naming:
  files:
    pattern: camelCase.js or kebab-case.js
    examples:
      - userService.js
      - auth-middleware.js

  directories:
    pattern: lowercase or kebab-case
    examples:
      - routes/
      - middleware/

  functions:
    pattern: camelCase
    examples:
      - getUserById
      - validateRequest

structure:
  new_route: |
    1. Create src/routes/{name}.js
    2. Add route handlers
    3. Register in src/routes/index.js

  new_service: |
    1. Create src/services/{name}Service.js
    2. Export service functions
    3. Add tests in __tests__/

errors:
  style: |
    Use async/await with try-catch.
    Create custom error classes for different error types.

testing:
  framework: Jest or Mocha
  pattern: "__tests__/{name}.test.js or {name}.spec.js"
  style: Use describe/it blocks with clear descriptions

commits:
  tool: Use checkpoint commit for all commits
  pre_commit: Run npm test before committing

rules:
  - Use async/await over callbacks
  - Validate all inputs
  - Use environment variables for config
  - Log errors with context

avoid:
  - Callback hell
  - Blocking the event loop
  - Exposing stack traces to clients
  - Hardcoded secrets

principles:
  - "Handle errors at boundaries"
  - "Validate early, fail fast"
  - "Keep middleware focused"
`

const nodeApiSkillsYml = `schema_version: "1"

global:
  - git
  - npm

local: []

config:
  npm:
    package_manager: npm
`
