import Navigation from "@/components/Navigation";
import HeroSection from "@/components/master-page/HeroSection";
import AboutSection from "@/components/master-page/AboutSection";
import HowItWorksSection from "@/components/master-page/HowItWorksSection";
import UseCasesSection from "@/components/master-page/UseCasesSection";
import GetStartedSection from "@/components/master-page/GetStartedSection";
import ContactSection from "@/components/master-page/ContactSection";
import Footer from "@/components/Footer";

export default function Home() {
  return (
    <main className="min-h-screen">
      <Navigation />
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
