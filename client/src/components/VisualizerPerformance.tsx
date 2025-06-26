import React, { useEffect, useState, useRef } from "react";
import { useRadio } from "../contexts/RadioContext";

export const VisualizerPerformance: React.FC = () => {
  const { isVisualizerEnabled, isPlaying } = useRadio();
  const [fps, setFps] = useState(0);
  const [frameCount, setFrameCount] = useState(0);
  const lastTimeRef = useRef<number>(0);
  const frameCountRef = useRef<number>(0);

  useEffect(() => {
    if (!isVisualizerEnabled || !isPlaying) {
      setFps(0);
      setFrameCount(0);
      return;
    }

    const updateFPS = () => {
      const now = performance.now();
      frameCountRef.current++;

      if (now - lastTimeRef.current >= 1000) {
        setFps(
          Math.round(
            (frameCountRef.current * 1000) / (now - lastTimeRef.current)
          )
        );
        setFrameCount(frameCountRef.current);
        frameCountRef.current = 0;
        lastTimeRef.current = now;
      }

      requestAnimationFrame(updateFPS);
    };

    lastTimeRef.current = performance.now();
    const animationId = requestAnimationFrame(updateFPS);

    return () => {
      cancelAnimationFrame(animationId);
    };
  }, [isVisualizerEnabled, isPlaying]);

  if (!isVisualizerEnabled) {
    return null;
  }

  return (
    <div className="fixed bottom-4 left-4 z-50 bg-black bg-opacity-80 border border-gray-700 p-3 rounded-sm">
      <div className="text-xs text-gray-400 font-mono mb-1">[PERFORMANCE]</div>
      <div className="text-xs font-mono">
        <div
          className={`${
            fps >= 55
              ? "text-green-400"
              : fps >= 30
              ? "text-yellow-400"
              : "text-red-400"
          }`}
        >
          FPS: {fps}
        </div>
        <div className="text-gray-500">Frames: {frameCount}</div>
      </div>
    </div>
  );
};
