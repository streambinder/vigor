export default {
  extends: ["@commitlint/config-conventional"],
  helpUrl: "https://www.conventionalcommits.org/",
  // until https://github.com/dependabot/dependabot-core/issues/2445 resolved
  ignores: [(msg) => /Signed-off-by: dependabot\[bot]/m.test(msg)],
};
