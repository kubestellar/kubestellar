import HeroSection from "@/components/master-page/HeroSection";
import AboutSection from "@/components/master-page/AboutSection";
import HowItWorksSection from "@/components/master-page/HowToUseSection";
import UseCasesSection from "@/components/master-page/UseCasesSection";
import GetStartedSection from "@/components/master-page/GetStartedSection";
import ContactSection from "@/components/master-page/ContactSection";
import { Navbar, Footer, PageActions } from "@/components";

export default function Home() {
  return (
    <main className="min-h-screen">
      <Navbar />
      <PageActions position="fixed" />
      <HeroSection />
      <AboutSection />
      <HowItWorksSection />
      <UseCasesSection />
      <GetStartedSection />
      <ContactSection />
      <Footer />
    </main>
  );
}
