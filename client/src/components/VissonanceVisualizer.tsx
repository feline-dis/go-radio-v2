import React, {
  useEffect,
  useRef,
  useCallback,
  useState,
} from "react";
import { useRadio } from "../contexts/RadioContext";

// Dynamic imports to handle ES module issues
let Vissonance: any;
let getPresets: any;

const initializeVissonance = async () => {
  if (!Vissonance) {
    const vissonanceModule = await import("vissonance");
    Vissonance = vissonanceModule.default || vissonanceModule;
    
    const presetsModule = await import("vissonance/presets");
    getPresets = presetsModule.getPresets || presetsModule.default?.getPresets;
    
    console.log("VissonanceVisualizer", Vissonance);
    console.log("presets", getPresets());
  }
};

// Import the Visualizer interface for proper typing
type Visualizer = {
  connectAudio(audioNode: AudioNode): void;
  loadPreset(preset: any, blendTime?: number): void;
  setRendererSize(width: number, height: number): void;
  render(): void;
  destroy(): void;
};

interface VissonanceVisualizerProps {
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

export const VissonanceVisualizer: React.FC<VissonanceVisualizerProps> = ({
  isEnabled,
  currentPreset,
  onPresetChange,
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const visualizerRef = useRef<Visualizer | null>(null);
  const isInitializedRef = useRef<boolean>(false);
  const resizeTimeoutRef = useRef<number | null>(null);
  const [webGLError, setWebGLError] = useState<string | null>(null);
  const { audioContextRef, gainNodeRef } = useRadio();

  // Initialize visualizer
  const initVisualizer = useCallback(async () => {
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
      // Initialize Vissonance modules
      await initializeVissonance();
      
      if (!Vissonance || !getPresets) {
        throw new Error("Failed to load Vissonance modules");
      }

      // Create visualizer with optimized settings
      const visualizer = Vissonance.createVisualizer(
        audioContextRef.current,
        canvasRef.current,
        {
          width: window.innerWidth,
          height: window.innerHeight,
        }
      );

      if (!visualizer) {
        throw new Error("Failed to create visualizer - createVisualizer returned null");
      }

      visualizerRef.current = visualizer;

      // Connect to audio
      visualizer.connectAudio(gainNodeRef.current);

      // Load initial preset
      const presets = getPresets();
      const presetNames = Object.keys(presets);
      
      if (presetNames.length > 0) {
        const presetName = currentPreset || presetNames[0];
        const preset = presets[presetName];
        if (preset) {
          console.log("visualizerRef.current:", visualizer);
          visualizer.loadPreset(preset, 0.0);
          if (!currentPreset) {
            onPresetChange(presetName);
          }
        }
      }

      isInitializedRef.current = true;
      setWebGLError(null);
      console.log("Vissonance visualizer initialized");
    } catch (error) {
      console.error("Failed to initialize Vissonance visualizer:", error);
      setWebGLError("Failed to initialize visualizer");
    }
  }, [
    audioContextRef,
    gainNodeRef,
    currentPreset,
    onPresetChange
  ]);

  // Handle preset change
  const handlePresetChange = useCallback(
    async (presetName: string) => {
      if (!visualizerRef.current) {
        return;
      }

      try {
        await initializeVissonance();
        const presets = getPresets();
        
        if (!presets[presetName]) {
          return;
        }

        visualizerRef.current.loadPreset(presets[presetName], 0.0);
        onPresetChange(presetName);
        console.log("Loaded preset:", presetName);
      } catch (error) {
        console.error("Failed to load preset:", error);
      }
    },
    [onPresetChange]
  );

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
    const initialize = async () => {
      if (
        isEnabled &&
        audioContextRef.current &&
        gainNodeRef.current &&
        !isInitializedRef.current
      ) {
        await initVisualizer();
      }
    };
    
    initialize();
  }, [isEnabled, audioContextRef, gainNodeRef, initVisualizer]);

  // Handle preset changes
  useEffect(() => {
    if (
      visualizerRef.current &&
      currentPreset &&
      isInitializedRef.current
    ) {
      handlePresetChange(currentPreset);
    }
  }, [currentPreset, handlePresetChange]);

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
      if (visualizerRef.current) {
        try {
          visualizerRef.current.destroy();
        } catch (error) {
          console.error("Error destroying visualizer:", error);
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
export const VissonancePresetSelector: React.FC<{
  currentPreset: string;
  onPresetChange: (preset: string) => void;
  isEnabled: boolean;
}> = ({ currentPreset, onPresetChange, isEnabled }) => {
  const [presetNames, setPresetNames] = useState<string[]>([]);

  useEffect(() => {
    const loadPresets = async () => {
      try {
        await initializeVissonance();
        if (getPresets) {
          const presets = getPresets();
          setPresetNames(Object.keys(presets));
        }
      } catch (error) {
        console.error("Failed to load presets for selector:", error);
      }
    };

    if (isEnabled) {
      loadPresets();
    }
  }, [isEnabled]);

  if (!isEnabled) {
    return null;
  }

  return (
    <div className="fixed top-20 right-4 z-50 bg-black bg-opacity-80 border border-gray-700 p-2 sm:p-3 rounded-sm">
      <div className="text-xs text-gray-400 font-mono mb-1 sm:mb-2 hidden sm:block">
        [VISSONANCE PRESET]
      </div>
      <select
        value={currentPreset}
        onChange={(e) => onPresetChange(e.target.value)}
        className="bg-black border border-gray-600 text-white text-xs font-mono px-1 sm:px-2 py-1 focus:outline-none focus:border-white w-full sm:w-auto"
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