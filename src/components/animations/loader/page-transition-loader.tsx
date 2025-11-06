"use client";

import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { Loader } from "@/components/animations/loader";

interface PageTransitionLoaderProps {
  children: React.ReactNode;
}

const PageTransitionLoader = ({ children }: PageTransitionLoaderProps) => {
  const [isLoading, setIsLoading] = useState(false);
  const pathname = usePathname();

  useEffect(() => {
    // Show loader when route changes
    const handleRouteChange = () => {
      setIsLoading(true);

      // Hide loader after content loads
      const timer = setTimeout(() => {
        setIsLoading(false);
      }, 400); // Reduced lag time

      return () => clearTimeout(timer);
    };

    // Handle Next.js route changes
    handleRouteChange();

    // Handle direct link clicks
    const handleLinkClick = (event: Event) => {
      const target = event.target as HTMLElement;
      const link = target.closest("a");

      if (link) {
        const href = link.getAttribute("href");
        const isExternal =
          link.getAttribute("target") === "_blank" ||
          href?.startsWith("http") ||
          href?.startsWith("//");

        // Show loader for internal navigation only
        if (!isExternal && href && href !== pathname) {
          setIsLoading(true);
        }
      }
    };

    // Handle browser back/forward buttons
    const handlePopState = () => {
      setIsLoading(true);
      setTimeout(() => setIsLoading(false), 300); // Reduced lag time
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
    }, 300); // Reduced lag time

    return () => clearTimeout(timer);
  }, [pathname]);

  return (
    <>
      <Loader isLoading={isLoading} text="Loading" />
      <div
        className={
          isLoading
            ? "opacity-0"
            : "opacity-100 transition-opacity duration-500"
        }
      >
        {children}
      </div>
    </>
  );
};

export default PageTransitionLoader;
