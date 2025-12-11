"use client";

import React from "react";
import EditViewSourceButtons from "./docs/EditViewSourceButtons";

interface PageActionsProps {
  filePath?: string;
  user?: string;
  repo?: string;
  branch?: string;
  docsPath?: string;
  position?: "top" | "fixed";
}

export default function PageActions({
  filePath,
  user,
  repo,
  branch,
  docsPath,
  position = "top",
}: PageActionsProps) {
  if (position === "fixed") {
    return (
      <div className="fixed top-20 right-4 z-40 hidden lg:block">
        <EditViewSourceButtons
          filePath={filePath}
          user={user}
          repo={repo}
          branch={branch}
          docsPath={docsPath}
        />
      </div>
    );
  }

  return (
    <EditViewSourceButtons
      filePath={filePath}
      user={user}
      repo={repo}
      branch={branch}
      docsPath={docsPath}
    />
  );
}
