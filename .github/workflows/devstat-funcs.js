const DEFAULT_REPOS = [
  "kubestellar/a2a",
  // 'kubestellar/community',
  // 'kubestellar/core',
  "kubestellar/docs",
  // 'kubestellar/galaxy',
  // 'kubestellar/helm',
  "kubestellar/homebrew-kubectl-multi",
  // 'kubestellar/homebrew-kubestellar',
  // 'kubestellar/infra',
  "kubestellar/kubectl-multi-plugin",
  "kubestellar/kubectl-rbac-flatten-plugin",
  "kubestellar/kubeflex",
  "kubestellar/kubestellar",
  "kubestellar/ocm-status-addon",
  "kubestellar/ocm-transport-plugin",
  // 'kubestellar/presentations',
  "kubestellar/ui",
  "kubestellar/ui-plugins",
];

function refreshSheet() {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getActiveSheet();
  // Insert a temporary row at the top
  sheet.insertRows(1, 1);
  // Delete the temporary row immediately
  sheet.deleteRow(1);
}

function getDateXDaysAgo(daysAgo) {
  const date = new Date();
  date.setDate(date.getDate() - daysAgo);
  return date.toISOString().split("T")[0]; // Format: YYYY-MM-DD
}

function buildGitHubQuery({
  username,
  repos,
  filters,
  qualifier,
  sinceDate,
  suffix = "",
  prefix = "",
}) {
  return `{${repos
    .map(repo => {
      const label = repo.split("/")[1].replace(/-/g, "_");
      const fieldName = `${prefix}${label}${suffix}`;
      const datePart = sinceDate ? ` created:>=${sinceDate}` : "";
      return `
      ${fieldName}: search(query: "${filters}${datePart} ${qualifier}:${username} repo:${repo}", type: ISSUE) { issueCount }
    `;
    })
    .join("\n")}}`;
}

function callGitHubGraphQL(query) {
  const response = UrlFetchApp.fetch("https://api.github.com/graphql", {
    method: "post",
    contentType: "application/json",
    payload: JSON.stringify({ query }),
    headers: {
      Authorization: `Bearer ${GITHUB_TOKEN}`,
    },
  });

  const data = JSON.parse(response.getContentText());

  if (data.errors) {
    throw new Error("GitHub API Error: " + JSON.stringify(data.errors));
  }

  return Object.values(data.data)
    .map(v => v?.issueCount || 0)
    .reduce((sum, count) => sum + count, 0);
}

// === OPEN PR COUNTS ===

function GET_OPEN_PR_COUNT_SINCE(username, sinceDate, repos = DEFAULT_REPOS) {
  const query = buildGitHubQuery({
    username,
    repos,
    filters: "is:pr is:open",
    qualifier: "author",
    sinceDate,
  });
  return callGitHubGraphQL(query);
}

function GET_OPEN_PR_COUNT_DYNAMIC(username, daysAgo, repos = DEFAULT_REPOS) {
  const sinceDate = getDateXDaysAgo(daysAgo);
  return GET_OPEN_PR_COUNT_SINCE(username, sinceDate, repos);
}

// === MERGED PR COUNTS ===

function GET_MERGED_PR_COUNT_SINCE(username, sinceDate, repos = DEFAULT_REPOS) {
  const query = buildGitHubQuery({
    username,
    repos,
    filters: "is:pr is:merged",
    qualifier: "author",
    sinceDate,
  });
  return callGitHubGraphQL(query);
}

function GET_MERGED_PR_COUNT_DYNAMIC(username, daysAgo, repos = DEFAULT_REPOS) {
  const sinceDate = getDateXDaysAgo(daysAgo);
  return GET_MERGED_PR_COUNT_SINCE(username, sinceDate, repos);
}

// === ASSIGNED ISSUE COUNTS ===

function GET_ASSIGNED_ISSUE_COUNT(username, repos = DEFAULT_REPOS) {
  const query = buildGitHubQuery({
    username,
    repos,
    filters: "is:issue is:open",
    qualifier: "assignee",
  });
  return callGitHubGraphQL(query);
}

function GET_ASSIGNED_ISSUE_COUNT_SINCE(
  username,
  sinceDate,
  repos = DEFAULT_REPOS
) {
  const query = buildGitHubQuery({
    username,
    repos,
    filters: "is:issue is:open",
    qualifier: "assignee",
    sinceDate,
  });
  return callGitHubGraphQL(query);
}

function GET_ASSIGNED_ISSUE_COUNT_DYNAMIC(
  username,
  daysAgo,
  repos = DEFAULT_REPOS
) {
  const sinceDate = getDateXDaysAgo(daysAgo);
  return GET_ASSIGNED_ISSUE_COUNT_SINCE(username, sinceDate, repos);
}

// === HELP WANTED ISSUE COUNTS ===

function GET_HELP_WANTED_COUNT_SINCE(
  username,
  sinceDate,
  repos = DEFAULT_REPOS
) {
  const query = buildGitHubQuery({
    username,
    repos,
    filters: 'is:issue label:\\"help wanted\\"',
    qualifier: "author",
    sinceDate,
  });
  return callGitHubGraphQL(query);
}

function GET_HELP_WANTED_COUNT_DYNAMIC(
  username,
  daysAgo,
  repos = DEFAULT_REPOS
) {
  const sinceDate = getDateXDaysAgo(daysAgo);
  return GET_HELP_WANTED_COUNT_SINCE(username, sinceDate, repos);
}

// === COMMENTED PR ACTIVITY (MERGED + OPEN) ===

function GET_COMMENTED_MERGED_AND_OPEN_PR_COUNT_SINCE(
  username,
  sinceDate,
  repos = DEFAULT_REPOS
) {
  const mergedQuery = buildGitHubQuery({
    username,
    repos,
    filters: "is:pr is:merged",
    qualifier: "commenter",
    sinceDate,
    suffix: "_merged",
  });
  const openQuery = buildGitHubQuery({
    username,
    repos,
    filters: "is:pr is:open",
    qualifier: "commenter",
    sinceDate,
    suffix: "_open",
  });

  const merged = callGitHubGraphQL(mergedQuery);
  const open = callGitHubGraphQL(openQuery);
  return merged + open;
}

function GET_COMMENTED_MERGED_AND_OPEN_PR_COUNT_DYNAMIC(
  username,
  daysAgo,
  repos = DEFAULT_REPOS
) {
  const sinceDate = getDateXDaysAgo(daysAgo);
  return GET_COMMENTED_MERGED_AND_OPEN_PR_COUNT_SINCE(
    username,
    sinceDate,
    repos
  );
}
