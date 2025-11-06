"use client";

import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import DocsLoader from "./loader";

interface DocsPageTransitionProps {
  children: React.ReactNode;
}

const DocsPageTransition = ({ children }: DocsPageTransitionProps) => {
  const [isLoading, setIsLoading] = useState(false);
  const pathname = usePathname();

  useEffect(() => {
    // Show loader when docs route changes
    const handleRouteChange = () => {
      setIsLoading(true);

      // Hide loader after content loads
      const timer = setTimeout(() => {
        setIsLoading(false);
      }, 300); // Reduced lag time

      return () => clearTimeout(timer);
    };

    // Handle docs navigation
    handleRouteChange();

    // Handle link clicks within docs
    const handleLinkClick = (event: Event) => {
      const target = event.target as HTMLElement;
      const link = target.closest("a");

      if (link) {
        const href = link.getAttribute("href");
        const isExternal =
          link.getAttribute("target") === "_blank" ||
          href?.startsWith("http") ||
          href?.startsWith("//");

        // Show loader for internal docs navigation only
        if (
          !isExternal &&
          href &&
          href !== pathname &&
          href.startsWith("/docs")
        ) {
          setIsLoading(true);
        }
      }
    };

    // Handle browser back/forward buttons
    const handlePopState = () => {
      if (window.location.pathname.startsWith("/docs")) {
        setIsLoading(true);
        setTimeout(() => setIsLoading(false), 200); // Reduced lag time
      }
    };

    // Add event listeners
    document.addEventListener("click", handleLinkClick);
    window.addEventListener("popstate", handlePopState);

    // Cleanup
    return () => {
      document.removeEventListener("click", handleLinkClick);
      window.removeEventListener("popstate", handlePopState);
    };
  }, [pathname]);

  // Hide loader when page is fully loaded
  useEffect(() => {
    const timer = setTimeout(() => {
      setIsLoading(false);
    }, 200); // Reduced lag time

    return () => clearTimeout(timer);
  }, [pathname]);

  return (
    <>
      <DocsLoader isLoading={isLoading} text="Loading Documentation" />
      <div
        className={
          isLoading
            ? "opacity-0"
            : "opacity-100 transition-opacity duration-300"
        }
      >
        {children}
      </div>
    </>
  );
};

export default DocsPageTransition;
