#!/usr/bin/env node
/**
 * Updates versions.ts with a new version entry
 *
 * Usage:
 *   node scripts/update-version.js --project kubestellar --version 0.30.0 --branch docs/0.30.0 [--set-latest]
 *
 * Options:
 *   --project      Project ID (kubestellar, a2a, kubeflex, multi-plugin)
 *   --version      Version number (e.g., 0.30.0)
 *   --branch       Branch name for this version (e.g., docs/0.30.0)
 *   --set-latest   If provided, updates the "latest" label, branch, and currentVersion
 */

const fs = require('fs');
const path = require('path');

// Parse command line arguments
const args = process.argv.slice(2);
const getArg = (name) => {
  const idx = args.indexOf(`--${name}`);
  return idx !== -1 && args[idx + 1] ? args[idx + 1] : null;
};

const project = getArg('project');
const version = getArg('version');
const branch = getArg('branch');
const setLatest = args.includes('--set-latest');

// Validate required arguments
if (!project || !version || !branch) {
  console.error('Usage: node update-version.js --project <project> --version <version> --branch <branch> [--set-latest]');
  console.error('');
  console.error('Required arguments:');
  console.error('  --project     Project ID (kubestellar, a2a, kubeflex, multi-plugin)');
  console.error('  --version     Version number (e.g., 0.30.0)');
  console.error('  --branch      Branch name (e.g., docs/0.30.0)');
  console.error('');
  console.error('Optional:');
  console.error('  --set-latest  Update "latest" label and currentVersion');
  process.exit(1);
}

// Project-specific version constant names
const versionConstants = {
  'kubestellar': 'KUBESTELLAR_VERSIONS',
  'a2a': 'A2A_VERSIONS',
  'kubeflex': 'KUBEFLEX_VERSIONS',
  'multi-plugin': 'MULTI_PLUGIN_VERSIONS',
  'kubectl-claude': 'KUBECTL_CLAUDE_VERSIONS'
};

const constName = versionConstants[project];
if (!constName) {
  console.error(`Unknown project: ${project}`);
  console.error(`Valid projects: ${Object.keys(versionConstants).join(', ')}`);
  process.exit(1);
}

// Escape special regex characters to prevent regex injection
function escapeRegex(str) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

// Read versions.ts
const versionsPath = path.join(__dirname, '../src/config/versions.ts');
if (!fs.existsSync(versionsPath)) {
  console.error(`File not found: ${versionsPath}`);
  process.exit(1);
}

let content = fs.readFileSync(versionsPath, 'utf8');
const originalContent = content;

console.log(`Updating ${project} to v${version}...`);
console.log(`  Branch: ${branch}`);
console.log(`  Set as latest: ${setLatest}`);

// Update latest label if setting as latest
let previousLatestVersion = null;
let previousLatestBranch = null;

if (setLatest) {
  // Extract current latest version before updating (to preserve as historical entry)
  const extractLabelRegex = new RegExp(
    `const ${constName}.*?latest:\\s*\\{.*?label:\\s*"v([\\d.]+) \\(Latest\\)"`,
    's'
  );
  const extractBranchRegex = new RegExp(
    `const ${constName}.*?latest:\\s*\\{.*?branch:\\s*"([^"]+)"`,
    's'
  );

  const labelMatch = content.match(extractLabelRegex);
  const branchMatch = content.match(extractBranchRegex);

  if (labelMatch && branchMatch) {
    previousLatestVersion = labelMatch[1];
    previousLatestBranch = branchMatch[1];
    // Only preserve if it's different from main (actual release)
    if (previousLatestBranch !== 'main' && previousLatestVersion !== version) {
      console.log(`  Previous latest: v${previousLatestVersion} (${previousLatestBranch})`);
    } else {
      previousLatestVersion = null;
      previousLatestBranch = null;
    }
  }

  // Update the "latest" entry's label to show the new version
  const latestRegex = new RegExp(
    `(const ${constName}.*?latest:\\s*\\{.*?label:\\s*")v[\\d.]+( \\(Latest\\)")`,
    's'
  );

  if (latestRegex.test(content)) {
    content = content.replace(latestRegex, `$1v${version}$2`);
    console.log(`  Updated latest label to v${version}`);
  } else {
    console.warn(`  Warning: Could not find latest label pattern in ${constName}`);
  }

  // Update the "latest" entry's branch to point to the frozen version branch
  const latestBranchRegex = new RegExp(
    `(const ${constName}.*?latest:\\s*\\{.*?branch:\\s*")[^"]+(")`,
    's'
  );

  if (latestBranchRegex.test(content)) {
    content = content.replace(latestBranchRegex, `$1${branch}$2`);
    console.log(`  Updated latest branch to ${branch}`);
  } else {
    console.warn(`  Warning: Could not find latest branch pattern in ${constName}`);
  }

  // Update currentVersion in the project config
  // Match the project section and update currentVersion
  const projectKey = project === 'multi-plugin' ? '"multi-plugin"' : project;
  const currentVersionRegex = new RegExp(
    `(${escapeRegex(projectKey)}:\\s*\\{.*?currentVersion:\\s*")([^"]+)(")`,
    's'
  );

  if (currentVersionRegex.test(content)) {
    content = content.replace(currentVersionRegex, `$1${version}$3`);
    console.log(`  Updated currentVersion to ${version}`);
  } else {
    console.warn(`  Warning: Could not find currentVersion for ${project}`);
  }
}

// Add version entry for previous latest (when setting new latest) or for non-latest versions
const versionEntryRegex = new RegExp(`"${escapeRegex(version)}":\\s*\\{`);
if (setLatest && previousLatestVersion) {
  // Add entry for the PREVIOUS latest version (so it remains accessible)
  const prevVersionEntryRegex = new RegExp(`"${escapeRegex(previousLatestVersion)}":\\s*\\{`);
  if (!prevVersionEntryRegex.test(content)) {
    const prevEntry = `  "${previousLatestVersion}": {
    label: "v${previousLatestVersion}",
    branch: "${previousLatestBranch}",
    isDefault: false,
  },`;

    const mainEntryRegex = new RegExp(
      `(const ${constName}.*?main:\\s*\\{[^}]+\\},)`,
      's'
    );

    if (mainEntryRegex.test(content)) {
      content = content.replace(mainEntryRegex, `$1\n${prevEntry}`);
      console.log(`  Added historical entry for previous latest v${previousLatestVersion}`);
    }
  } else {
    console.log(`  Previous latest v${previousLatestVersion} already has an entry`);
  }
} else if (setLatest) {
  console.log(`  No previous latest to preserve (was pointing to main)`);
} else if (versionEntryRegex.test(content)) {
  console.log(`  Version ${version} already exists in ${constName}, skipping addition`);
} else {
  // Add new version entry after 'main' entry
  const newEntry = `  "${version}": {
    label: "v${version}",
    branch: "${branch}",
    isDefault: false,
  },`;

  // Find the main entry in the specific version constant and add after it
  const mainEntryRegex = new RegExp(
    `(const ${constName}.*?main:\\s*\\{[^}]+\\},)`,
    's'
  );

  if (mainEntryRegex.test(content)) {
    content = content.replace(mainEntryRegex, `$1\n${newEntry}`);
    console.log(`  Added version entry for ${version}`);
  } else {
    // Fallback: try to add after latest entry
    const latestEntryRegex = new RegExp(
      `(const ${constName}.*?latest:\\s*\\{[^}]+\\},)`,
      's'
    );

    if (latestEntryRegex.test(content)) {
      content = content.replace(latestEntryRegex, `$1\n${newEntry}`);
      console.log(`  Added version entry after latest for ${version}`);
    } else {
      console.error(`  Error: Could not find insertion point for new version`);
      process.exit(1);
    }
  }
}

// Write updated content
if (content !== originalContent) {
  fs.writeFileSync(versionsPath, content);
  console.log(`\n✅ Updated ${versionsPath}`);
} else {
  console.log(`\nNo changes needed for ${versionsPath}`);
}

// Also update shared.json for dynamic version loading
const sharedJsonPath = path.join(__dirname, '../public/config/shared.json');
if (fs.existsSync(sharedJsonPath)) {
  console.log('\nUpdating shared.json...');
  const sharedConfig = JSON.parse(fs.readFileSync(sharedJsonPath, 'utf8'));

  // Initialize project versions if not exists
  if (!sharedConfig.versions[project]) {
    sharedConfig.versions[project] = {};
  }

  // Capture previous latest before updating (for shared.json)
  let sharedPrevVersion = null;
  let sharedPrevBranch = null;
  if (setLatest && sharedConfig.versions[project].latest) {
    const prevLabel = sharedConfig.versions[project].latest.label;
    sharedPrevBranch = sharedConfig.versions[project].latest.branch;
    // Extract version from label like "v0.4.4 (Latest)"
    const match = prevLabel.match(/^v([\d.]+)/);
    if (match && sharedPrevBranch !== 'main' && match[1] !== version) {
      sharedPrevVersion = match[1];
      console.log(`  Previous latest: v${sharedPrevVersion} (${sharedPrevBranch})`);
    }
  }

  // Update latest label and branch if setting as latest
  if (setLatest && sharedConfig.versions[project].latest) {
    sharedConfig.versions[project].latest.label = `v${version} (Latest)`;
    sharedConfig.versions[project].latest.branch = branch;
    console.log(`  Updated latest label to v${version}`);
    console.log(`  Updated latest branch to ${branch}`);
  }

  // Update currentVersion in projects
  if (setLatest && sharedConfig.projects && sharedConfig.projects[project]) {
    sharedConfig.projects[project].currentVersion = version;
    console.log(`  Updated currentVersion to ${version}`);
  }

  // Add entry for previous latest (when setting new latest) or for non-latest versions
  if (setLatest && sharedPrevVersion && !sharedConfig.versions[project][sharedPrevVersion]) {
    // Add entry for the PREVIOUS latest version (so it remains accessible)
    sharedConfig.versions[project][sharedPrevVersion] = {
      label: `v${sharedPrevVersion}`,
      branch: sharedPrevBranch,
      isDefault: false
    };
    console.log(`  Added historical entry for previous latest v${sharedPrevVersion}`);
  } else if (setLatest && sharedPrevVersion) {
    console.log(`  Previous latest v${sharedPrevVersion} already has an entry`);
  } else if (setLatest) {
    console.log(`  No previous latest to preserve (was pointing to main)`);
  } else if (!sharedConfig.versions[project][version]) {
    sharedConfig.versions[project][version] = {
      label: `v${version}`,
      branch: branch,
      isDefault: false
    };
    console.log(`  Added version entry for ${version}`);
  }

  // Update timestamp
  sharedConfig.updatedAt = new Date().toISOString();

  // Write updated shared.json
  fs.writeFileSync(sharedJsonPath, JSON.stringify(sharedConfig, null, 2) + '\n');
  console.log(`✅ Updated ${sharedJsonPath}`);
} else {
  console.log(`\nWarning: ${sharedJsonPath} not found, skipping shared config update`);
}
