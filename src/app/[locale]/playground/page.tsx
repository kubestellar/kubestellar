"use client";

import { useEffect, useState } from "react";
import { useRouter } from "@/i18n/navigation";
import { Loader } from "@/components/animations/loader";

export default function PlaygroundPage() {
  const router = useRouter();
  //eslint-disable-next-line @typescript-eslint/no-unused-vars 
  const [isRedirecting, setIsRedirecting] = useState(true);

  useEffect(() => {
    // Redirect immediately to coming-soon page
    router.replace("/coming-soon");
  }, [router]);

  return <Loader isLoading={isRedirecting} text="Redirecting to Coming Soon" />;
}
