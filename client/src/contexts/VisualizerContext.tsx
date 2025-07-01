import { createContext, useContext, useState } from "react";
import type { ReactNode } from "react";

interface VisualizerContextType {
  isVisualizerEnabled: boolean;
  setIsVisualizerEnabled: (enabled: boolean) => void;
  currentPreset: string;
  setCurrentPreset: (preset: string) => void;
}

const VisualizerContext = createContext<VisualizerContextType | undefined>(undefined);

export const useVisualizer = () => {
  const context = useContext(VisualizerContext);
  if (context === undefined) {
    throw new Error("useVisualizer must be used within a VisualizerProvider");
  }
  return context;
};

export const VisualizerProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [isVisualizerEnabled, setIsVisualizerEnabled] = useState(false);
  const [currentPreset, setCurrentPreset] = useState("");

  const value: VisualizerContextType = {
    isVisualizerEnabled,
    setIsVisualizerEnabled,
    currentPreset,
    setCurrentPreset,
  };

  return (
    <VisualizerContext.Provider value={value}>
      {children}
    </VisualizerContext.Provider>
  );
}; 