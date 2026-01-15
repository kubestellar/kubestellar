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
 *   --set-latest   If provided, updates the "latest" label and currentVersion
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
if (setLatest) {
  // Update the "latest" entry's label to show the new version
  const latestRegex = new RegExp(
    `(const ${constName}[\\s\\S]*?latest:\\s*\\{[\\s\\S]*?label:\\s*")v[\\d.]+( \\(Latest\\)")`,
    ''
  );

  if (latestRegex.test(content)) {
    content = content.replace(latestRegex, `$1v${version}$2`);
    console.log(`  Updated latest label to v${version}`);
  } else {
    console.warn(`  Warning: Could not find latest label pattern in ${constName}`);
  }

  // Update currentVersion in the project config
  // Match the project section and update currentVersion
  const projectKey = project === 'multi-plugin' ? '"multi-plugin"' : project;
  const currentVersionRegex = new RegExp(
    `(${escapeRegex(projectKey)}:\\s*\\{[\\s\\S]*?currentVersion:\\s*")([^"]+)(")`,
    ''
  );

  if (currentVersionRegex.test(content)) {
    content = content.replace(currentVersionRegex, `$1${version}$3`);
    console.log(`  Updated currentVersion to ${version}`);
  } else {
    console.warn(`  Warning: Could not find currentVersion for ${project}`);
  }
}

// Check if version entry already exists
const versionEntryRegex = new RegExp(`"${escapeRegex(version)}":\\s*\\{`, '');
if (versionEntryRegex.test(content)) {
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
    `(const ${constName}[\\s\\S]*?main:\\s*\\{[^}]+\\},)`,
    ''
  );

  if (mainEntryRegex.test(content)) {
    content = content.replace(mainEntryRegex, `$1\n${newEntry}`);
    console.log(`  Added version entry for ${version}`);
  } else {
    // Fallback: try to add after latest entry
    const latestEntryRegex = new RegExp(
      `(const ${constName}[\\s\\S]*?latest:\\s*\\{[^}]+\\},)`,
      ''
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
