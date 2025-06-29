import React, {
  useEffect,
  useRef,
  useCallback,
  useMemo,
  useState,
} from "react";
import butterchurn from "butterchurn";
import type { Visualizer } from "butterchurn";
import butterchurnPresets from "butterchurn-presets";
import { useRadio } from "../contexts/RadioContext";

interface ButterchurnVisualizerProps {
  isEnabled: boolean;
  currentPreset: string;
  onPresetChange: (preset: string) => void;
}

// WebGL support detection
const isWebGLSupported = (): boolean => {
  try {
    const canvas = document.createElement("canvas");
    const gl = canvas.getContext("webgl2") || canvas.getContext("webgl");
    return !!gl;
  } catch {
    return false;
  }
};

export const ButterchurnVisualizer: React.FC<ButterchurnVisualizerProps> = ({
  isEnabled,
  currentPreset,
  onPresetChange,
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const visualizerRef = useRef<Visualizer | null>(null);
  const animationFrameRef = useRef<number | null>(null);
  const isInitializedRef = useRef<boolean>(false);
  const resizeTimeoutRef = useRef<number | null>(null);
  const [webGLError, setWebGLError] = useState<string | null>(null);
  const { audioContextRef, gainNodeRef, isPlaying } = useRadio();

  // Memoize presets to avoid recreating on every render
  const { presets, presetNames } = useMemo(() => {
    const presets = butterchurnPresets.getPresets();
    return {
      presets,
      presetNames: Object.keys(presets),
    };
  }, []);

  // Initialize visualizer
  const initVisualizer = useCallback(() => {
    if (
      !canvasRef.current ||
      !audioContextRef.current ||
      !gainNodeRef.current ||
      isInitializedRef.current
    ) {
      return;
    }

    // Check WebGL support
    if (!isWebGLSupported()) {
      setWebGLError("WebGL is not supported in this browser");
      return;
    }

    try {
      // Create visualizer with optimized settings
      visualizerRef.current = butterchurn.createVisualizer(
        audioContextRef.current,
        canvasRef.current,
        {
          width: 600,
          height: 600,
        }
      );
      visualizerRef.current.setRendererSize(600, 600);

      // Connect to audio
      visualizerRef.current.connectAudio(gainNodeRef.current);

      // Load initial preset
      if (presetNames.length > 0) {
        const presetName = currentPreset || presetNames[0];
        const preset = presets[presetName];
        if (preset) {
          visualizerRef.current.loadPreset(preset, 0.0);
          if (!currentPreset) {
            onPresetChange(presetName);
          }
        }
      }

      isInitializedRef.current = true;
      setWebGLError(null);
      console.log("Butterchurn visualizer initialized");
    } catch (error) {
      console.error("Failed to initialize Butterchurn visualizer:", error);
      setWebGLError("Failed to initialize visualizer");
    }
  }, [
    audioContextRef,
    gainNodeRef,
    currentPreset,
    presetNames,
    presets,
    onPresetChange,
  ]);

  // Handle preset change
  const handlePresetChange = useCallback(
    (presetName: string) => {
      if (!visualizerRef.current || !presets[presetName]) {
        return;
      }

      try {
        visualizerRef.current.loadPreset(presets[presetName], 0.0);
        onPresetChange(presetName);
        console.log("Loaded preset:", presetName);
      } catch (error) {
        console.error("Failed to load preset:", error);
      }
    },
    [presets, onPresetChange]
  );

  // Optimized animation loop with throttling
  const renderFrame = useCallback(() => {
    if (visualizerRef.current && isEnabled && isPlaying) {
      try {
        visualizerRef.current.render();
      } catch (error) {
        console.error("Error rendering visualizer:", error);
        // Stop rendering on error to prevent infinite error loops
        if (animationFrameRef.current) {
          cancelAnimationFrame(animationFrameRef.current);
          animationFrameRef.current = null;
        }
      }
    }
    // Use a more efficient animation frame request
    animationFrameRef.current = requestAnimationFrame(renderFrame);
  }, [isEnabled, isPlaying]);

  // Debounced resize handler
  const handleResize = useCallback(() => {
    if (!visualizerRef.current || !canvasRef.current) return;

    // Debounce resize to avoid excessive calls
    if (resizeTimeoutRef.current) {
      clearTimeout(resizeTimeoutRef.current);
    }
    resizeTimeoutRef.current = window.setTimeout(() => {
      if (visualizerRef.current) {
        try {
          visualizerRef.current.setRendererSize(
            window.innerWidth,
            window.innerHeight
          );
        } catch (error) {
          console.error("Error resizing visualizer:", error);
        }
      }
    }, 100);
  }, []);

  // Initialize visualizer when audio context is ready
  useEffect(() => {
    if (
      isEnabled &&
      audioContextRef.current &&
      gainNodeRef.current &&
      !isInitializedRef.current
    ) {
      initVisualizer();
    }
  }, [isEnabled, audioContextRef, gainNodeRef, initVisualizer]);

  // Handle preset changes
  useEffect(() => {
    if (
      visualizerRef.current &&
      currentPreset &&
      presets[currentPreset] &&
      isInitializedRef.current
    ) {
      handlePresetChange(currentPreset);
    }
  }, [currentPreset, presets, handlePresetChange]);

  // Start/stop animation loop with better cleanup
  useEffect(() => {
    if (isEnabled && visualizerRef.current && isInitializedRef.current) {
      renderFrame();
    } else if (animationFrameRef.current) {
      cancelAnimationFrame(animationFrameRef.current);
      animationFrameRef.current = null;
    }

    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
        animationFrameRef.current = null;
      }
    };
  }, [isEnabled, renderFrame]);

  // Handle window resize with cleanup
  useEffect(() => {
    window.addEventListener("resize", handleResize);
    return () => {
      window.removeEventListener("resize", handleResize);
      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current);
        resizeTimeoutRef.current = null;
      }
    };
  }, [handleResize]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
        animationFrameRef.current = null;
      }
      if (visualizerRef.current) {
        try {
          visualizerRef.current.disconnect();
        } catch (error) {
          console.error("Error disconnecting visualizer:", error);
        }
        visualizerRef.current = null;
      }
      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current);
        resizeTimeoutRef.current = null;
      }
      isInitializedRef.current = false;
    };
  }, []);

  if (!isEnabled) {
    return null;
  }

  if (webGLError) {
    return (
      <div className="fixed inset-0 z-0 pointer-events-none flex items-center justify-center">
        <div className="bg-black bg-opacity-80 border border-gray-700 p-4 rounded-sm">
          <div className="text-xs text-red-400 font-mono">
            [VISUALIZER ERROR: {webGLError}]
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 z-0 pointer-events-none">
      <canvas
        ref={canvasRef}
        className="w-full h-full"
        style={{
          display: "block",
        }}
      />
    </div>
  );
};

// Optimized preset selector component
export const PresetSelector: React.FC<{
  currentPreset: string;
  onPresetChange: (preset: string) => void;
  isEnabled: boolean;
}> = ({ currentPreset, onPresetChange, isEnabled }) => {
  // Memoize presets to avoid recreating on every render
  const { presetNames } = useMemo(() => {
    const presets = butterchurnPresets.getPresets();
    return {
      presetNames: Object.keys(presets),
    };
  }, []);

  if (!isEnabled) {
    return null;
  }

  return (
    <div className="fixed top-4 right-4 z-50 bg-black bg-opacity-80 border border-gray-700 p-3 rounded-sm">
      <div className="text-xs text-gray-400 font-mono mb-2">
        [VISUALIZER PRESET]
      </div>
      <select
        value={currentPreset}
        onChange={(e) => onPresetChange(e.target.value)}
        className="bg-black border border-gray-600 text-white text-xs font-mono px-2 py-1 focus:outline-none focus:border-white"
      >
        {presetNames.map((presetName) => (
          <option key={presetName} value={presetName}>
            {presetName}
          </option>
        ))}
      </select>
    </div>
  );
};
