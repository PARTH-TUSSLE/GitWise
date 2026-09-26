"use client";

import React, { useEffect, useRef } from "react";

interface NodePoint {
  x: number;
  y: number;
  baseX: number;
  baseY: number;
  label?: string;
  depth: number;
  pulseTimer: number;
  illumination: number;
  isSecondary: boolean;
}

interface GraphEdge {
  from: number;
  to: number;
  progress: number;
  hasSignal: boolean;
  signalSpeed: number;
  signalColorType: "primary" | "secondary";
}

export function CompilerBackground() {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d", { alpha: true });
    if (!ctx) return;

    let animationFrameId: number;
    let width = window.innerWidth;
    let height = window.innerHeight;

    const setupCanvas = () => {
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      width = window.innerWidth;
      height = window.innerHeight;
      canvas.width = Math.floor(width * dpr);
      canvas.height = Math.floor(height * dpr);
      canvas.style.width = `${width}px`;
      canvas.style.height = `${height}px`;
      ctx.setTransform(1, 0, 0, 1, 0, 0);
      ctx.scale(dpr, dpr);
    };

    setupCanvas();

    // Check prefers-reduced-motion
    let prefersReducedMotion = false;
    try {
      prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    } catch {
      prefersReducedMotion = false;
    }

    // Pointer coordinates with smooth easing
    let mouseX = width / 2;
    let mouseY = height / 3;
    let targetMouseX = width / 2;
    let targetMouseY = height / 3;
    let hasPointerMoved = false;

    const handlePointerMove = (e: MouseEvent) => {
      targetMouseX = e.clientX;
      targetMouseY = e.clientY;
      hasPointerMoved = true;
    };

    window.addEventListener("mousemove", handlePointerMove, { passive: true });

    // CSS Theme variables extractor (Dual-Tone Support)
    const getThemeColors = () => {
      if (typeof window === "undefined") {
        return {
          accent: "#f43f5e",
          accentRgb: "244, 63, 94",
          accentSecondary: "#06b6d4",
          accentSecondaryRgb: "6, 182, 212",
          borderRgb: "45, 45, 45",
        };
      }
      const rootStyle = getComputedStyle(document.documentElement);
      const rawAccent = rootStyle.getPropertyValue("--accent").trim() || "#f43f5e";
      const rawAccentRgb = rootStyle.getPropertyValue("--accent-rgb").trim() || "244, 63, 94";
      const rawSecondary = rootStyle.getPropertyValue("--accent-secondary").trim() || "#06b6d4";
      const rawSecondaryRgb = rootStyle.getPropertyValue("--accent-secondary-rgb").trim() || "6, 182, 212";
      const rawBorderRgb = rootStyle.getPropertyValue("--border-rgb").trim() || "45, 45, 45";

      const accentRgb = rawAccentRgb.split(",").length === 3 ? rawAccentRgb : "244, 63, 94";
      const accentSecondaryRgb = rawSecondaryRgb.split(",").length === 3 ? rawSecondaryRgb : "6, 182, 212";
      const borderRgb = rawBorderRgb.split(",").length === 3 ? rawBorderRgb : "45, 45, 45";

      return {
        accent: rawAccent,
        accentRgb,
        accentSecondary: rawSecondary,
        accentSecondaryRgb,
        borderRgb,
      };
    };

    let theme = getThemeColors();

    const onThemeChange = () => {
      theme = getThemeColors();
    };

    window.addEventListener("gitwise-theme-change", onThemeChange);

    const observer = new MutationObserver((mutations) => {
      for (const m of mutations) {
        if (m.type === "attributes" && m.attributeName === "data-theme") {
          onThemeChange();
        }
      }
    });

    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["data-theme"],
    });

    // Topological nodes positioned across negative space
    let nodes: NodePoint[] = [];
    let edges: GraphEdge[] = [];
    let scanlineY = 0;

    const initGraph = () => {
      nodes = [];
      edges = [];

      const labels = [
        "AST:ROOT", "CORE:INGEST", "SWC:PARSER", "DAG:TRACE",
        "PKG:SERVER", "ACTION:DISPATCH", "FLIGHT:STREAM", "CRYPTO:HMAC",
        "K8S:INFORMER", "TURBO:CACHE", "SCHED:LANE", "API:CONTRACT",
        "DIFF:SYNTH", "AST:NODE_42", "MEM:ZERO_COPY", "SOCKET:ACK",
        "DRAIN:CTRL", "TREE:SITTER", "WASM:INDEX", "TRACE:CALL"
      ];

      // Coordinate anchors across viewport margins and hero shoulders
      const leftCol = Math.max(28, width * 0.05);
      const leftMid = Math.max(80, width * 0.12);
      const rightCol = width - Math.max(28, width * 0.05);
      const rightMid = width - Math.max(80, width * 0.12);

      const points = [
        // Left Column AST Spine (Hero negative space)
        { x: leftCol, y: height * 0.10, label: labels[0], depth: 1, isSecondary: false },
        { x: leftMid, y: height * 0.20, label: labels[1], depth: 2, isSecondary: true },
        { x: leftCol + 40, y: height * 0.35, label: labels[2], depth: 3, isSecondary: false },
        { x: leftMid + 20, y: height * 0.52, label: labels[3], depth: 2, isSecondary: true },
        { x: leftCol + 30, y: height * 0.72, label: labels[4], depth: 4, isSecondary: false },
        { x: leftMid - 20, y: height * 0.88, label: labels[16], depth: 3, isSecondary: true },

        // Right Column AST Spine (Hero negative space)
        { x: rightCol, y: height * 0.12, label: labels[5], depth: 1, isSecondary: true },
        { x: rightMid, y: height * 0.24, label: labels[6], depth: 3, isSecondary: false },
        { x: rightCol - 40, y: height * 0.40, label: labels[7], depth: 2, isSecondary: true },
        { x: rightMid - 30, y: height * 0.58, label: labels[8], depth: 4, isSecondary: false },
        { x: rightCol - 20, y: height * 0.76, label: labels[9], depth: 3, isSecondary: true },
        { x: rightMid + 10, y: height * 0.90, label: labels[17], depth: 2, isSecondary: false },

        // Upper Hero Shoulders (flanking headline)
        { x: width * 0.22, y: height * 0.14, label: labels[10], depth: 2, isSecondary: false },
        { x: width * 0.78, y: height * 0.15, label: labels[11], depth: 2, isSecondary: true },

        // Intermediate section bridge points
        { x: width * 0.18, y: height * 0.64, label: labels[12], depth: 3, isSecondary: true },
        { x: width * 0.82, y: height * 0.66, label: labels[13], depth: 3, isSecondary: false },
        { x: width * 0.28, y: height * 0.94, label: labels[18], depth: 4, isSecondary: false },
        { x: width * 0.72, y: height * 0.95, label: labels[19], depth: 4, isSecondary: true },
      ];

      nodes = points.map((pt) => ({
        x: pt.x,
        y: pt.y,
        baseX: pt.x,
        baseY: pt.y,
        label: pt.label,
        depth: pt.depth,
        pulseTimer: Math.random() * 6.28,
        illumination: 0.2,
        isSecondary: pt.isSecondary,
      }));

      // Orthogonal Manhattan connections
      const connections: [number, number, "primary" | "secondary"][] = [
        [0, 1, "primary"], [1, 2, "secondary"], [2, 3, "primary"], [3, 4, "secondary"], [4, 5, "primary"],
        [6, 7, "secondary"], [7, 8, "primary"], [8, 9, "secondary"], [9, 10, "primary"], [10, 11, "secondary"],
        [0, 12, "primary"], [12, 1, "secondary"],
        [6, 13, "secondary"], [13, 7, "primary"],
        [2, 14, "secondary"], [14, 3, "primary"],
        [8, 15, "primary"], [15, 9, "secondary"],
        [4, 16, "secondary"], [10, 17, "primary"],
      ];

      edges = connections
        .filter(([from, to]) => from < nodes.length && to < nodes.length)
        .map(([from, to, type], i) => ({
          from,
          to,
          progress: (i * 0.14) % 1,
          hasSignal: true,
          signalSpeed: 0.0022 + (i % 4) * 0.0012,
          signalColorType: type,
        }));
    };

    initGraph();

    const handleResize = () => {
      setupCanvas();
      initGraph();
    };

    window.addEventListener("resize", handleResize);

    // Main animation loop
    const render = () => {
      // Easing for pointer
      if (hasPointerMoved) {
        mouseX += (targetMouseX - mouseX) * 0.08;
        mouseY += (targetMouseY - mouseY) * 0.08;
      }

      ctx.clearRect(0, 0, width, height);

      // 1. Faint Compiler Grid
      const gridSize = 64;
      ctx.beginPath();
      ctx.strokeStyle = `rgba(${theme.borderRgb}, 0.35)`;
      ctx.lineWidth = 0.5;

      for (let x = 0; x < width; x += gridSize) {
        ctx.moveTo(x, 0);
        ctx.lineTo(x, height);
      }
      for (let y = 0; y < height; y += gridSize) {
        ctx.moveTo(0, y);
        ctx.lineTo(width, y);
      }
      ctx.stroke();

      // 2. Compiler Grid Dual-Tone Crosshairs (+) at key junctions
      const crossSize = 3.5;
      const crossStep = gridSize * 3;
      let tickToggle = 0;

      for (let x = crossStep; x < width; x += crossStep) {
        for (let y = crossStep; y < height; y += crossStep) {
          tickToggle++;
          const colorRgb = tickToggle % 2 === 0 ? theme.accentRgb : theme.accentSecondaryRgb;
          ctx.fillStyle = `rgba(${colorRgb}, 0.32)`;
          ctx.fillRect(x - crossSize, y - 0.5, crossSize * 2, 1);
          ctx.fillRect(x - 0.5, y - crossSize, 1, crossSize * 2);
        }
      }

      // 3. Faint Telemetry Scanline Sweep
      if (!prefersReducedMotion) {
        scanlineY = (scanlineY + 0.8) % height;
        const grad = ctx.createLinearGradient(0, scanlineY - 30, 0, scanlineY + 30);
        grad.addColorStop(0, "rgba(0, 0, 0, 0)");
        grad.addColorStop(0.5, `rgba(${theme.accentRgb}, 0.04)`);
        grad.addColorStop(1, "rgba(0, 0, 0, 0)");
        ctx.fillStyle = grad;
        ctx.fillRect(0, scanlineY - 30, width, 60);
      }

      // 4. Parallax Node Positions & Proximity Illumination
      const parallaxFactor = 0.008;
      nodes.forEach((node) => {
        if (!prefersReducedMotion && hasPointerMoved) {
          const offsetX = (mouseX - width / 2) * parallaxFactor * (1 / node.depth);
          const offsetY = (mouseY - height / 2) * parallaxFactor * (1 / node.depth);
          node.x = node.baseX + offsetX;
          node.y = node.baseY + offsetY;
        } else {
          node.x = node.baseX;
          node.y = node.baseY;
        }

        node.pulseTimer += 0.03;
        const breathing = (Math.sin(node.pulseTimer) + 1) * 0.5; // 0..1

        // Proximity calculation (mouse distance)
        const dist = Math.hypot(mouseX - node.x, mouseY - node.y);
        const proximity = Math.max(0, 1 - dist / 220);
        node.illumination = 0.15 + breathing * 0.1 + proximity * 0.6;
      });

      // 5. Draw Orthogonal Edges (Manhattan L-routing) with Dual-Tone Lighting
      edges.forEach((edge) => {
        const fromNode = nodes[edge.from];
        const toNode = nodes[edge.to];
        if (!fromNode || !toNode) return;

        const cornerX = toNode.x;
        const cornerY = fromNode.y;

        const edgeIllumination = Math.max(fromNode.illumination, toNode.illumination);
        const edgeAlpha = 0.12 + edgeIllumination * 0.28;
        const edgeColorRgb = edge.signalColorType === "primary" ? theme.accentRgb : theme.accentSecondaryRgb;

        ctx.beginPath();
        ctx.strokeStyle = `rgba(${edgeColorRgb}, ${edgeAlpha})`;
        ctx.lineWidth = 1;
        ctx.moveTo(fromNode.x, fromNode.y);
        ctx.lineTo(cornerX, cornerY);
        ctx.lineTo(toNode.x, toNode.y);
        ctx.stroke();

        // 6. Traveling Signal Packets (Dual-Tone Zero-radius square data frames)
        if (edge.hasSignal) {
          if (!prefersReducedMotion) {
            edge.progress += edge.signalSpeed;
            if (edge.progress > 1) edge.progress = 0;
          }

          let px = fromNode.x;
          let py = fromNode.y;

          if (edge.progress < 0.5) {
            const t = edge.progress * 2;
            px = fromNode.x + (cornerX - fromNode.x) * t;
            py = fromNode.y;
          } else {
            const t = (edge.progress - 0.5) * 2;
            px = cornerX;
            py = cornerY + (toNode.y - cornerY) * t;
          }

          // Main signal packet (visible solid square in contrasting color)
          const packetAlpha = 0.6 + edgeIllumination * 0.4;
          const signalRgb = edge.signalColorType === "primary" ? theme.accentRgb : theme.accentSecondaryRgb;

          ctx.fillStyle = `rgba(${signalRgb}, ${packetAlpha})`;
          ctx.fillRect(px - 2, py - 2, 4, 4);

          // Faint glow contrail
          ctx.fillStyle = `rgba(${signalRgb}, ${packetAlpha * 0.3})`;
          ctx.fillRect(px - 4, py - 4, 8, 8);
        }
      });

      // 7. Draw Nodes & Technical Monospace Labels with Dual-Tone Framing
      nodes.forEach((node) => {
        const nodeRgb = node.isSecondary ? theme.accentSecondaryRgb : theme.accentRgb;

        // Core block
        ctx.fillStyle = `rgba(${nodeRgb}, ${0.35 + node.illumination * 0.5})`;
        ctx.fillRect(node.x - 2, node.y - 2, 4, 4);

        // Technical outer frame bracket
        ctx.strokeStyle = `rgba(${nodeRgb}, ${0.28 + node.illumination * 0.5})`;
        ctx.lineWidth = 1;
        ctx.strokeRect(node.x - 5, node.y - 5, 10, 10);

        // Micro-Label in negative space
        if (node.label) {
          ctx.font = '10px ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace';
          ctx.fillStyle = `rgba(180, 180, 180, ${0.45 + node.illumination * 0.45})`;
          ctx.fillText(node.label, node.x + 9, node.y + 3.5);
        }
      });

      if (!prefersReducedMotion) {
        animationFrameId = requestAnimationFrame(render);
      }
    };

    render();

    return () => {
      cancelAnimationFrame(animationFrameId);
      window.removeEventListener("mousemove", handlePointerMove);
      window.removeEventListener("resize", handleResize);
      window.removeEventListener("gitwise-theme-change", onThemeChange);
      observer.disconnect();
    };
  }, []);

  return (
    <div
      aria-hidden="true"
      className="fixed inset-0 pointer-events-none z-0 overflow-hidden"
    >
      <canvas
        ref={canvasRef}
        className="w-full h-full block"
        style={{
          width: "100%",
          height: "100%",
        }}
      />
    </div>
  );
}
