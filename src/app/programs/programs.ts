export interface Program {
  id: string;
  name: string;
  fullName: string;
  description: string;
  logo: string;
  isPaid: boolean;
  theme: {
    gradient: string;
    primaryColor: string;
    secondaryColor: string;
    floatingShapes: string[];
  };
  sections: {
    benefits: string;
    description: string;
    overview: string;
    eligibility: string;
    timeline: string;
    structure: string;
    howToApply: string;
    resources: Array<{
      name: string;
      url: string;
    }>;
  };
}

export const programs: Program[] = [
  {
    id: "gsoc",
    name: "GSoC",
    fullName: "Google Summer of Code",
    description:
      "Transform your coding skills with Google's premier open source program",
    logo: "/GSoC image.png",
    isPaid: true,
    theme: {
      gradient: "linear-gradient(135deg, #FDE047, #FACC15, #EAB308, #CA8A04)",
      primaryColor: "#FDE047",
      secondaryColor: "#CA8A04",
      floatingShapes: ["bg-yellow-500", "bg-yellow-400", "bg-amber-500"],
    },
    sections: {
      benefits:
        "Gain real-world experience, learn from experienced mentors, become part of an open-source community, and receive a stipend upon successful completion of the program.",
      description:
        "Google Summer of Code is a global program focused on bringing more student developers into open source software development. Students work with an open source organization on a 3-month programming project during their break from school.",
      overview:
        "KubeStellar participates as a mentor organization. We provide project ideas, mentors, and a welcoming community for students to learn and contribute.",
      eligibility:
        "Participants must be at least 18 years of age, a student or a beginner to open source software development. For detailed criteria, please refer to the official GSoC website.",
      timeline:
        "The program typically runs from May to August. Key dates include the application period, community bonding, and coding phases. Check the GSoC website for the official timeline.",
      structure:
        "Accepted contributors work on their project with the guidance of one or more mentors from KubeStellar. There are evaluations at the midpoint and end of the program.",
      howToApply:
        "Students can apply via the Google Summer of Code website during the application period. We recommend engaging with our community and contributing to KubeStellar beforehand.",
      resources: [
        {
          name: "Official GSoC Website",
          url: "https://summerofcode.withgoogle.com/",
        },
        {
          name: "KubeStellar GitHub",
          url: "https://github.com/kubestellar/kubestellar",
        },
      ],
    },
  },
  {
    id: "esoc",
    name: "ESoC",
    fullName: "European Summer of Code",
    description: "Empower European talent in open source development",
    logo: "/ESoC.png",
    isPaid: true,
    theme: {
      gradient: "linear-gradient(135deg, #CA8A04, #A16207, #92400E, #78350F)",
      primaryColor: "#CA8A04",
      secondaryColor: "#78350F",
      floatingShapes: ["bg-yellow-700", "bg-amber-800", "bg-yellow-800"],
    },
    sections: {
      benefits:
        "A great opportunity to work on a real-world project, receive a stipend, and connect with the open source community.",
      description:
        "European Summer of Code is a program aimed at students and recent graduates in Europe, providing them with an opportunity to contribute to open source projects.",
      overview:
        "KubeStellar is excited to mentor participants in ESoC, offering challenging projects and dedicated support to help them grow as developers.",
      eligibility:
        "Open to students and recent graduates based in Europe. Please check the official ESoC website for detailed eligibility rules.",
      timeline:
        "The program usually takes place during the summer months. Refer to the ESoC website for the specific dates and deadlines.",
      structure:
        "Participants work closely with KubeStellar mentors on a pre-defined project, with regular check-ins and feedback sessions.",
      howToApply:
        "Applications should be submitted through the official European Summer of Code portal.",
      resources: [
        {
          name: "Official ESoC Website (Link not available)",
          url: "#",
        },
        {
          name: "KubeStellar GitHub",
          url: "https://github.com/kubestellar/kubestellar",
        },
      ],
    },
  },
  {
    id: "ifos",
    name: "IFoS",
    fullName: "Interns for Open Source",
    description: "Kickstart your open source journey with KubeStellar",
    logo: "/IFoS logo.png",
    isPaid: false,
    theme: {
      gradient: "linear-gradient(135deg, #06B6D4, #0891B2, #0E7490, #155E75)",
      primaryColor: "#06B6D4",
      secondaryColor: "#155E75",
      floatingShapes: ["bg-cyan-500", "bg-teal-500", "bg-cyan-600"],
    },
    sections: {
      benefits:
        "While this is an unpaid program, successful graduates receive a certificate of completion, a letter of recommendation, and are given priority consideration for any future paid mentorship programs like GSoC or LFX.",
      description:
        "Interns for Open Source (IFoS) is a unique, unpaid internship program created by KubeStellar. It is designed for individuals who are passionate about open source and want to gain hands-on experience with a cutting-edge project.",
      overview:
        "This 3-month program provides a direct pathway into the KubeStellar community. Participants work on meaningful projects and are mentored by our core developers.",
      eligibility:
        "We welcome applications from anyone with a strong interest in Kubernetes, multi-cluster orchestration, and open source. Basic knowledge of Go and container technologies is a plus.",
      timeline:
        "IFoS is a rolling program. Applications are accepted year-round, and internships start based on project availability and applicant schedules.",
      structure:
        "Interns are paired with a mentor and integrated into one of our development teams. The program is flexible, allowing for part-time or full-time commitment.",
      howToApply:
        "To apply, please send your resume and a brief statement of interest to our community email. We also encourage you to start contributing to our GitHub repository.",
      resources: [
        {
          name: "KubeStellar GitHub",
          url: "https://github.com/kubestellar/kubestellar",
        },
        {
          name: "KubeStellar Community Page",
          url: "https://kubestellar.io/community",
        },
      ],
    },
  },
  {
    id: "lfx",
    name: "LFX",
    fullName: "LFX Mentorship",
    description:
      "Accelerate your open source journey with Linux Foundation mentorship",
    logo: "/lfx-logo.png",
    isPaid: true,
    theme: {
      gradient: "linear-gradient(135deg, #3B82F6, #1D4ED8, #1E40AF, #1E3A8A)",
      primaryColor: "#3B82F6",
      secondaryColor: "#1E3A8A",
      floatingShapes: ["bg-blue-500", "bg-blue-400", "bg-blue-600"],
    },
    sections: {
      benefits:
        "Receive a stipend, gain hands-on experience with cutting-edge technology, build your professional network, and enhance your resume.",
      description:
        "The LFX Mentorship program, run by the Linux Foundation, provides a structured, remote learning opportunity for aspiring open source contributors. Mentees get to work on real-world projects with experienced mentors.",
      overview:
        "KubeStellar is proud to be a part of the LFX Mentorship program. We offer projects that are critical to our roadmap, giving mentees a chance to make a significant impact.",
      eligibility:
        "The program is open to developers from all backgrounds. Specific requirements may vary by project. Check the LFX Mentorship platform for details on eligibility.",
      timeline:
        "LFX Mentorship runs in terms, typically Spring, Summer, and Fall. Each term is about 12 weeks long.",
      structure:
        "Mentees work one-on-one with a mentor from KubeStellar, contributing to the project and participating in the community.",
      howToApply:
        "Applications are submitted through the LFX Mentorship platform. Browse for KubeStellar projects and apply.",
      resources: [
        {
          name: "LFX Mentorship Platform",
          url: "https://mentorship.lfx.linuxfoundation.org/",
        },
        {
          name: "KubeStellar GitHub",
          url: "https://github.com/kubestellar/kubestellar",
        },
      ],
    },
  },
];

export function getProgramById(id: string): Program | undefined {
  return programs.find(program => program.id === id);
}

export function getAllPrograms(): Program[] {
  return programs;
}
