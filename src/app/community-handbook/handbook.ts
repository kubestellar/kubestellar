// Community Handbook data structure
export interface HandbookCard {
  id: string;
  title: string;
  description: string;
  iconType: string;
  iconPath: string;
  bgColor: string;
  iconColor: string;
  link: string;
}

export const handbookCards: HandbookCard[] = [
  {
    id: "onboarding",
    title: "Onboarding",
    description: "KubeStellar GitHub Organization On-boarding and Off-boarding Policy. Learn how to get started with our community.",
    iconType: "user-plus",
    iconPath: "M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z",
    bgColor: "bg-blue-500/20",
    iconColor: "text-blue-400",
    link: "/docs/contribution-guidelines/onboarding"
  },
  {
    id: "code-of-conduct",
    title: "Code of Conduct",
    description: "Our pledge to create a welcoming and inclusive community for everyone to contribute and thrive.",
    iconType: "shield-check",
    iconPath: "M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z",
    bgColor: "bg-purple-500/20",
    iconColor: "text-purple-400",
    link: "/docs/contribution-guidelines/code-of-conduct"
  },
  {
    id: "guidelines",
    title: "Guidelines",
    description: "Best practices for contributing to the KubeStellar project. Essential guidelines for quality contributions.",
    iconType: "document-text",
    iconPath: "M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z",
    bgColor: "bg-emerald-500/20",
    iconColor: "text-emerald-400",
    link: "/docs/contribution-guidelines/contributing"
  },
  {
    id: "license",
    title: "License",
    description: "KubeStellar is licensed under the Apache 2.0 License. Learn about open source licensing and terms.",
    iconType: "lock-closed",
    iconPath: "M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z",
    bgColor: "bg-yellow-500/20",
    iconColor: "text-yellow-400",
    link: "/docs/contribution-guidelines/license"
  },
  {
    id: "governance",
    title: "Governance",
    description: "How the KubeStellar project is governed and organized. Understand our decision-making processes.",
    iconType: "users",
    iconPath: "M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z",
    bgColor: "bg-cyan-500/20",
    iconColor: "text-cyan-400",
    link: "/docs/contribution-guidelines/governance"
  },
  {
    id: "testing",
    title: "Testing",
    description: "Procedures and guidelines for testing contributions. Ensure quality and reliability in every change.",
    iconType: "check-circle",
    iconPath: "M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z",
    bgColor: "bg-red-500/20",
    iconColor: "text-red-400",
    link: "/docs/contribution-guidelines/testing"
  },
  {
    id: "packaging",
    title: "Packaging",
    description: "How to package and distribute KubeStellar components. Learn about build and deployment processes.",
    iconType: "cube",
    iconPath: "M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4",
    bgColor: "bg-indigo-500/20",
    iconColor: "text-indigo-400",
    link: "/docs/contribution-guidelines/packaging"
  },
  {
    id: "docs-management",
    title: "Docs Management Overview",
    description: "Overview of how documentation is managed and updated. Comprehensive documentation workflow.",
    iconType: "book-open",
    iconPath: "M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.746 0 3.332.477 4.5 1.253v13C19.832 18.477 18.246 18 16.5 18c-1.746 0-3.332.477-4.5 1.253",
    bgColor: "bg-teal-500/20",
    iconColor: "text-teal-400",
    link: "/docs/contribution-guidelines/docs-management"
  },
  {
    id: "testing-website-prs",
    title: "Testing Website PRs",
    description: "How to test pull requests for the KubeStellar website. Quality assurance for web changes.",
    iconType: "check-circle-2",
    iconPath: "M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z",
    bgColor: "bg-green-500/20",
    iconColor: "text-green-400",
    link: "/docs/contribution-guidelines/testing-website-prs"
  },
  {
    id: "release-process",
    title: "Release Process",
    description: "The process for creating and publishing new KubeStellar releases. Complete release lifecycle management.",
    iconType: "flag",
    iconPath: "M7 4V2a1 1 0 011-1h8a1 1 0 011 1v2m0 0V3a1 1 0 011 1v8.5l-5-5-5 5V4a1 1 0 011-1m0 0h10m-9 4h8m-8 4h8m-8 4h3.5",
    bgColor: "bg-orange-500/20",
    iconColor: "text-orange-400",
    link: "/docs/contribution-guidelines/release-process"
  },
  {
    id: "release-testing",
    title: "Release Testing",
    description: "How to test and validate new releases before publication. Comprehensive release validation process.",
    iconType: "check-circle-3",
    iconPath: "M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z",
    bgColor: "bg-pink-500/20",
    iconColor: "text-pink-400",
    link: "/docs/contribution-guidelines/release-testing"
  },
  {
    id: "signoff-signing",
    title: "Signoff and Signing Contributions",
    description: "Requirements for signing off on your contributions. Legal compliance and contribution verification.",
    iconType: "pencil",
    iconPath: "M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z",
    bgColor: "bg-violet-500/20",
    iconColor: "text-violet-400",
    link: "/docs/contribution-guidelines/signoff-signing"
  }
];
