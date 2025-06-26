import React from "react";
import { useRadio } from "../contexts/RadioContext";

export const VisualizerToggle: React.FC = () => {
  const { isVisualizerEnabled, setIsVisualizerEnabled, isAudioContextReady } =
    useRadio();

  const handleToggle = () => {
    setIsVisualizerEnabled(!isVisualizerEnabled);
  };

  if (!isAudioContextReady) {
    return null;
  }

  return (
    <div className="fixed top-4 left-4 z-50 bg-black bg-opacity-80 border border-gray-700 p-3 rounded-sm">
      <div className="text-xs text-gray-400 font-mono mb-2">[VISUALIZER]</div>
      <button
        onClick={handleToggle}
        className={`px-3 py-1 text-xs font-mono border transition-colors ${
          isVisualizerEnabled
            ? "bg-white text-black border-white"
            : "bg-black text-white border-gray-600 hover:border-white"
        }`}
      >
        {isVisualizerEnabled ? "[ON]" : "[OFF]"}
      </button>
    </div>
  );
};
