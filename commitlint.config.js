/**
 * @file Commitlint configuration
 * @description Enforces conventional commit format: type(scope?): subject
 *
 * @example
 * feat(benchmark): add RFC 2544 frame loss test
 * fix(reflector): resolve UDP packet drop
 * docs: update API reference
 * chore(deps): upgrade gopacket
 */
export default {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "type-enum": [
      2,
      "always",
      [
        "feat",     // New feature
        "fix",      // Bug fix
        "docs",     // Documentation changes
        "style",    // Code style changes (formatting)
        "refactor", // Code refactoring
        "perf",     // Performance improvements
        "test",     // Adding or updating tests
        "chore",    // Maintenance tasks
        "ci",       // CI/CD changes
        "build",    // Build system changes
        "revert",   // Revert a previous commit
      ],
    ],
    "scope-enum": [
      1,
      "always",
      [
        // Backend components
        "api", "auth", "config", "license", "testmaster",
        // Test modules (per stem's module architecture)
        "benchmark", "servicetest", "trafficgen", "measure", "certify", "reflector",
        // C dataplane
        "dataplane",
        // Frontend
        "ui", "components", "hooks",
        // Infrastructure
        "deps", "ci", "docker", "release",
      ],
    ],
    "subject-case": [2, "never", ["start-case", "pascal-case", "upper-case"]],
    "subject-full-stop": [2, "never", "."],
    "subject-empty": [2, "never"],
    "type-case": [2, "always", "lower-case"],
    "type-empty": [2, "never"],
    "header-max-length": [2, "always", 100],
  },
};
