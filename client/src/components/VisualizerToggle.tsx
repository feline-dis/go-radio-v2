import React from "react";
import { useRadio } from "../contexts/RadioContext";
import { useVisualizer } from "../contexts/VisualizerContext";

export const VisualizerToggle: React.FC = () => {
  const { isAudioContextReady } = useRadio();
  const { isVisualizerEnabled, setIsVisualizerEnabled } = useVisualizer();

  const handleToggle = () => {
    setIsVisualizerEnabled(!isVisualizerEnabled);
  };

  if (!isAudioContextReady) {
    return null;
  }

  return (
    <div className="fixed top-20 left-4 z-50 bg-black bg-opacity-80 border border-gray-700 p-2 sm:p-3 rounded-sm">
      <div className="text-xs text-gray-400 font-mono mb-1 sm:mb-2 hidden sm:block">[VISUALIZER]</div>
      <button
        onClick={handleToggle}
        className={`px-2 sm:px-3 py-1 text-xs font-mono border transition-colors ${
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
