import Navigation from "@/components/Navigation";
import HeroSection from "@/components/landing-page/HeroSection";
import AboutSection from "@/components/landing-page/AboutSection";
import HowItWorksSection from "@/components/landing-page/HowItWorksSection";
import UseCasesSection from "@/components/landing-page/UseCasesSection";
import GetStartedSection from "@/components/landing-page/GetStartedSection";
import ContactSection from "@/components/landing-page/ContactSection";
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
