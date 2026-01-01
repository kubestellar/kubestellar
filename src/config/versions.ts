// Versions config - docs are now stored locally in this repository
// Version information is kept for display purposes only
// 
// Versioning Strategy:
// - Docs are stored in this repository
// - Branch naming convention: docs/{version} (e.g., docs/0.29.0, docs/0.28.0)
// - The main branch always contains the latest version
// - To add a new version:
//   1. Create a new branch named docs/{version} from main
//   2. Add the version to VERSIONS below
//   3. Update CURRENT_VERSION if it's the latest

export const CURRENT_VERSION = "0.29.0"

// Available versions - branch name is derived from version key
// Branch naming: main for latest, docs/{version} for specific versions
export const VERSIONS = {
  latest: {
    label: `v${CURRENT_VERSION} (Latest)`,
    branch: "main",
    isDefault: true,
  },
  // Example of how to add older versions:
  // "0.28.0": {
  //   label: "v0.28.0",
  //   branch: "docs/0.28.0",
  //   isDefault: false,
  // },
} as const

export type VersionKey = keyof typeof VERSIONS

export interface VersionInfo {
  label: string
  branch: string
  isDefault: boolean
}

export function getDefaultVersion(): VersionKey {
  return "latest"
}

export function getCurrentVersion(): string {
  return CURRENT_VERSION
}

export function getBranchForVersion(version: VersionKey): string {
  return VERSIONS[version]?.branch ?? "main"
}

export function getVersionFromBranch(branch: string): VersionKey | null {
  // Check if branch matches docs/{version} pattern
  const match = branch.match(/^docs\/(.+)$/)
  if (match) {
    const versionNum = match[1]
    // Find version entry with matching branch
    for (const [key, value] of Object.entries(VERSIONS)) {
      if (value.branch === branch || key === versionNum) {
        return key as VersionKey
      }
    }
  }
  
  // Check for main branch
  if (branch === "main" || branch === "master") {
    return "latest"
  }
  
  return null
}

export function getAllVersions(): Array<{ key: VersionKey } & VersionInfo> {
  return Object.entries(VERSIONS).map(([key, value]) => ({
    key: key as VersionKey,
    ...value,
  }))
}

// Helper to validate if a branch name follows version convention
export function isVersionBranch(branch: string): boolean {
  return branch === "main" || branch.startsWith("docs/")
}
