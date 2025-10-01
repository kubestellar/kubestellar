import { getProgramById } from "../programs";
import { notFound } from "next/navigation";
import ProgramPageClient from "./ProgramPageClient";

interface PageProps {
  params: Promise<{
    slug: string;
  }>;
}

export default async function ProgramPage({ params }: PageProps) {
  const { slug } = await params;
  const program = getProgramById(slug);

  if (!program) {
    notFound();
  }

  return <ProgramPageClient program={program} />;
}
