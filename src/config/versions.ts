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

// Netlify site name for branch deploys
// Note: Branch deploys must be enabled in Netlify dashboard:
// Site Settings > Build & Deploy > Branches > Branch deploys: All
export const NETLIFY_SITE_NAME = "ks"

// Production URL for latest version
export const PRODUCTION_URL = "https://kubestellar.io"

// Available versions - branch name is derived from version key
// Branch naming: main for latest, docs/{version} for specific versions
export const VERSIONS = {
  latest: {
    label: `v${CURRENT_VERSION} (Latest)`,
    branch: "main",
    isDefault: true,
  },
  "0.28.0": {
    label: "v0.28.0",
    branch: "docs/0.28.0",
    isDefault: false,
  },
  "0.27.2": {
    label: "v0.27.2",
    branch: "docs/0.27.2",
    isDefault: false,
  },
  "0.27.1": {
    label: "v0.27.1",
    branch: "docs/0.27.1",
    isDefault: false,
  },
  "0.27.0": {
    label: "v0.27.0",
    branch: "docs/0.27.0",
    isDefault: false,
  },
  "0.26.0": {
    label: "v0.26.0",
    branch: "docs/0.26.0",
    isDefault: false,
  },
  "0.25.1": {
    label: "v0.25.1",
    branch: "docs/0.25.1",
    isDefault: false,
  },
  "0.25.0": {
    label: "v0.25.0",
    branch: "docs/0.25.0",
    isDefault: false,
  },
  "0.24.0": {
    label: "v0.24.0",
    branch: "docs/0.24.0",
    isDefault: false,
  },
  "0.23.1": {
    label: "v0.23.1",
    branch: "docs/0.23.1",
    isDefault: false,
  },
  "0.23.0": {
    label: "v0.23.0",
    branch: "docs/0.23.0",
    isDefault: false,
  },
  "0.22.0": {
    label: "v0.22.0",
    branch: "docs/0.22.0",
    isDefault: false,
  },
  "0.21.2": {
    label: "v0.21.2",
    branch: "docs/0.21.2",
    isDefault: false,
  },
  "0.21.1": {
    label: "v0.21.1",
    branch: "docs/0.21.1",
    isDefault: false,
  },
  "0.21.0": {
    label: "v0.21.0",
    branch: "docs/0.21.0",
    isDefault: false,
  },
  legacy: {
    label: "Older Versions",
    branch: "legacy",
    isDefault: false,
    externalUrl: "https://kubestellar.github.io/kubestellar",
  },
} as const

export type VersionKey = keyof typeof VERSIONS

export interface VersionInfo {
  label: string
  branch: string
  isDefault: boolean
  externalUrl?: string
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

// Get the URL for a specific version
export function getVersionUrl(versionKey: VersionKey, pathname: string = "/docs"): string {
  const version = VERSIONS[versionKey]

  // If it has an external URL (like legacy), use that
  if ('externalUrl' in version && version.externalUrl) {
    return version.externalUrl
  }

  // Latest version uses production URL
  if (versionKey === "latest" || version.isDefault) {
    return `${PRODUCTION_URL}${pathname}`
  }

  // For older versions, use the GitHub Pages legacy site
  // This provides immediate access to all historical versions
  // TODO: Switch to Netlify branch deploys once enabled in dashboard
  const versionNumber = versionKey.toString()
  return `https://kubestellar.github.io/kubestellar/release-${versionNumber}/`
}

// Check if a version has been migrated (branch exists)
export function isVersionMigrated(versionKey: VersionKey): boolean {
  // Latest is always available
  if (versionKey === "latest") return true

  // Legacy links externally, so it's "available"
  if (versionKey === "legacy") return true

  // For other versions, assume they exist if in the VERSIONS list
  // In practice, you might want to check if the branch actually exists
  return versionKey in VERSIONS
}
