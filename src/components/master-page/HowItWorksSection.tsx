"use client";

import { useEffect } from "react";

export default function HowItWorksSection() {
    useEffect(() => {
        // Create starfield and grid for How It Works section
        const createStarfield = (container: HTMLElement) => {
            if (!container) return;
            container.innerHTML = "";

            for (let layer = 1; layer <= 3; layer++) {
                const layerDiv = document.createElement("div");
                layerDiv.className = `star-layer layer-${layer}`;
                layerDiv.style.position = "absolute";
                layerDiv.style.inset = "0";
                layerDiv.style.zIndex = layer.toString();

                const starCount = layer === 1 ? 60 : layer === 2 ? 40 : 25;

                for (let i = 0; i < starCount; i++) {
                    const star = document.createElement("div");
                    star.style.position = "absolute";
                    star.style.width = `${Math.random() * 2 + 1}px`;
                    star.style.height = star.style.width;
                    star.style.backgroundColor = "white";
                    star.style.borderRadius = "50%";
                    star.style.top = `${Math.random() * 100}%`;
                    star.style.left = `${Math.random() * 100}%`;
                    star.style.opacity = Math.random().toString();
                    star.style.animation = `twinkle ${Math.random() * 3 + 2}s infinite alternate`;
                    star.style.animationDelay = `${Math.random() * 2}s`;
                    layerDiv.appendChild(star);
                }

                container.appendChild(layerDiv);
            }
        };

        const createGrid = (container: HTMLElement) => {
            if (!container) return;
            container.innerHTML = "";

            const gridSvg = document.createElementNS(
                "http://www.w3.org/2000/svg",
                "svg"
            );
            gridSvg.setAttribute("width", "100%");
            gridSvg.setAttribute("height", "100%");
            gridSvg.style.position = "absolute";
            gridSvg.style.top = "0";
            gridSvg.style.left = "0";

            for (let i = 0; i < 12; i++) {
                const line = document.createElementNS(
                    "http://www.w3.org/2000/svg",
                    "line"
                );
                line.setAttribute("x1", "0");
                line.setAttribute("y1", `${i * 8}%`);
                line.setAttribute("x2", "100%");
                line.setAttribute("y2", `${i * 8}%`);
                line.setAttribute("stroke", "#6366F1");
                line.setAttribute("stroke-width", "0.5");
                line.setAttribute("stroke-opacity", "0.3");
                gridSvg.appendChild(line);
            }

            for (let i = 0; i < 12; i++) {
                const line = document.createElementNS(
                    "http://www.w3.org/2000/svg",
                    "line"
                );
                line.setAttribute("x1", `${i * 8}%`);
                line.setAttribute("y1", "0");
                line.setAttribute("x2", `${i * 8}%`);
                line.setAttribute("y2", "100%");
                line.setAttribute("stroke", "#6366F1");
                line.setAttribute("stroke-width", "0.5");
                line.setAttribute("stroke-opacity", "0.3");
                gridSvg.appendChild(line);
            }

            container.appendChild(gridSvg);
        };

        const starsContainer = document.getElementById("stars-container-how");
        const gridContainer = document.getElementById("grid-lines-how");

        if (starsContainer) createStarfield(starsContainer);
        if (gridContainer) createGrid(gridContainer);
    }, []);

    return (
        <section
            id="how-it-works"
            className="relative py-16 bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white overflow-hidden"
        >
            {/* Dark base background */}
            <div className="absolute inset-0 bg-[#0a0a0a]"></div>

            {/* Starfield background */}
            <div
                id="stars-container-how"
                className="absolute inset-0 overflow-hidden"
            ></div>

            {/* Grid lines background */}
            <div id="grid-lines-how" className="absolute inset-0 opacity-20"></div>

            <div className="absolute right-0 top-0 h-full w-1/2 bg-gradient-to-l from-blue-500/10 to-transparent"></div>
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
                <div className="text-center mb-16">
                    <h2 className="text-3xl font-extrabold text-white sm:text-4xl">
                        How <span className="text-gradient">It Works</span>
                    </h2>
                    <p className="mt-4 max-w-2xl mx-auto text-xl text-gray-300">
                        KubeStellar orchestrates your multi-cluster environment with a simple, <br />
                        powerful architecture
                    </p>
                </div>

                {/* Interactive Workflow Steps */}
                <div className="relative z-10">
                    {/* Connection Line */}
                    <div className="absolute left-1/2 top-0 bottom-0 w-0.5 bg-gradient-to-b from-blue-500 to-purple-600 hidden md:block z-5"></div>

                    {/* Step 1 */}
                    <div className="relative mb-16 z-20">
                        <div className="flex flex-col md:flex-row items-center">
                            <div className="md:w-1/2 md:pr-10">
                                <div className="relative bg-gray-800/40 backdrop-blur-md rounded-s-md rounded-e-md p-4 border-1 border-white/10 p-6 z-30">
                                    <h3 className="text-xl font-bold  text-white mb-4 flex items-center">
                                        <span className="flex items-center justify-center w-10 h-10 rounded-full bg-blue-600 mr-2 text-white font-bold">
                                            1
                                        </span>

                                        Define Workloads
                                    </h3>
                                    <p className="text-gray-300 mb-6 leading-relaxed">
                                        Create Kubernetes resource in the KubeStellar control plane using familiar tools and manifests. Tag resource with placement constraints and policies.
                                    </p>
                                    <div className="bg-slate-900/90 rounded-lg overflow-hidden shadow-lg w-fit">
                                        {/* Code block */}
                                        <pre className="text-sm font-mono text-white p-4 leading-6">
                                            <code>
                                                <span className="text-yellow-300">apiVersion</span>: <span className="text-white">apps/v1</span>
                                                {"\n"}<span className="text-yellow-300">kind</span>: <span className="text-white">Deployment</span>
                                                {"\n"}<span className="text-yellow-300">metadata</span>:
                                                {"\n"}  <span className="text-yellow-300">name</span>: <span className="text-white">example-app</span>
                                                {"\n"}  <span className="text-yellow-300">annotations</span>:
                                                {"\n"}    <span className="text-yellow-300">kubestellar.io/placement</span>: <span className="text-emerald-400">&quot;region=us-east,tier=prod&quot;</span>
                                            </code>
                                        </pre>
                                    </div>

                                </div>
                            </div>

                            <div className="flex justify-center md:w-1/2 md:pl-12 mt-8 md:mt-0">
                                <div className="relative w-24 h-24 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center shadow-lg z-30">
                                    <svg
                                        className="w-12 h-12 text-white"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                    >
                                        <path
                                            strokeLinecap="round"
                                            strokeLinejoin="round"
                                            strokeWidth={2}
                                            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                                        />
                                    </svg>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Step 2 */}
                    <div className="relative mb-16 z-20">
                        <div className="flex flex-col md:flex-row-reverse items-center">
                            <div className="md:w-1/2 md:pl-12">
                                <div className="relative bg-gray-800/40 backdrop-blur-md rounded-s-md rounded-e-md p-4 border-1 border-white/10 z-30">

                                    <h3 className="text-xl font-bold  text-white mb-4 flex items-center">
                                        <span className="flex items-center justify-center w-10 h-10 rounded-full bg-blue-600 mr-2 text-white font-bold">
                                            2
                                        </span>

                                        Workload Orchestation
                                    </h3>

                                    <p className="text-gray-300 mb-6 leading-relaxed">
                                        KubeStellar&apos;s orchestation engine analyzes workloads and determines optimal placement across registered clusters based on constraints and policies.
                                    </p>
                                    <div className="grid grid-cols-3 gap-4">
                                        <div className="bg-blue-900/80 backdrop-blur-lg rounded-full px-3 py-1 text-white text-sm flex items-center justify-center">
                                            <div className="text-sm text-400 opacity-70 font-semibold">Policy Evaluation</div>
                                        </div>
                                        <div className="bg-purple-900/80 backdrop-blur-lg rounded-full px-3 py-1 text-white text-sm flex items-center justify-center">
                                            <div className="text-sm text-400 opacity-70 font-semibold">Constraint Matching</div>
                                        </div>
                                        <div className="bg-green-900/80 backdrop-blur-lg rounded-full px-3 py-1 text-white text-sm flex items-center justify-center">
                                            <div className="text-sm text-400 opacity-70 font-semibold">Resource Analysis</div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <div className="flex justify-center md:w-1/2 md:pr-12 mt-8 md:mt-0">
                                <div className="relative w-24 h-24 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center shadow-lg z-30">
                                    <svg
                                        className="w-12 h-12 text-white"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                    >
                                        <path
                                            strokeLinecap="round"
                                            strokeLinejoin="round"
                                            strokeWidth={1}
                                            d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                                        />
                                        <path
                                            strokeLinecap="round"
                                            strokeLinejoin="round"
                                            strokeWidth={1}
                                            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                                        />
                                    </svg>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Step 3 */}
                    <div className="relative z-20">
                        <div className="flex flex-col md:flex-row items-center">
                            <div className="md:w-1/2 md:pr-12">
                                <div className="relative bg-gray-800/40 backdrop-blur-md rounded-s-md rounded-e-md p-4 border-1 border-white/10 z-30">


                                    <h3 className="text-xl font-bold  text-white mb-4 flex items-center">
                                        <span className="flex items-center justify-center w-10 h-10 rounded-full bg-blue-600 mr-2 text-white font-bold">
                                            3
                                        </span>
                                        Automated Deployment
                                    </h3>


                                    <p className="text-gray-300 mb-6 leading-relaxed">
                                        Workloads are automatically deployed to selected clusters
                                        KubeStellar continuously monitors health and ensures desired state across all clusters.
                                    </p>
                                    <div className="flex items-center space-x-4">
                                        <div className="bg-blue-900/40 backdrop-blur-lg px-3 py-2 text-white text-sm flex flex-col items-center justify-center w-40 rounded-s rounded-e">
                                            <span className="text-sm text-400 opacity-50">Edge Cluster</span>

                                            <div className="w-full h-1 bg-blue-500 mt-2"></div>
                                        </div>

                                        <div className="bg-purple-900/40 backdrop-blur-lg px-3 py-2 text-white text-sm flex flex-col items-center justify-center w-40 rounded-s rounded-e">
                                            <span className="text-sm text-400 opacity-50">Cloud Cluster</span>
                                            <div className="w-full h-1 bg-purple-500 mt-2"></div>
                                        </div>

                                        <div className="bg-green-900/40 backdrop-blur-lg px-3 py-2 text-white text-sm flex flex-col items-center justify-center w-40 rounded-s rounded-e">
                                            <span className="text-sm text-400 opacity-50">On-Prem Cluster</span>
                                            <div className="w-full h-1 bg-green-500 mt-2"></div>
                                        </div>
                                    </div>



                                </div>
                            </div>
                            <div className="flex justify-center md:w-1/2 md:pl-12 mt-8 md:mt-0">
                                <div className="relative w-24 h-24 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center shadow-lg z-30">
                                    <svg
                                        className="w-12 h-12 text-white"
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                    >
                                        <path
                                            strokeLinecap="round"
                                            strokeLinejoin="round"
                                            strokeWidth={2}
                                            d="M7 16a4 4 0 01-.88-7.906A6 6 0 1118 14h-1m-5 6V10m0 10l-3-3m3 3l3-3"
                                        />
                                    </svg>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </section>
    );
}