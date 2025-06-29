declare module 'vissonance' {
  interface VisualizerOptions {
    width?: number;
    height?: number;
  }

  interface Preset {
    name: string;
    description?: string;
    class: any;
  }

  interface Visualizer {
    connectAudio(audioNode: AudioNode): void;
    loadPreset(preset: Preset, blendTime?: number): void;
    setRendererSize(width: number, height: number): void;
    render(): void;
    destroy(): void;
  }

  class Vissonance {
    constructor(audioContext: AudioContext, canvas: HTMLCanvasElement, options?: VisualizerOptions);
    connectAudio(audioNode: AudioNode): void;
    loadPreset(preset: Preset, blendTime?: number): void;
    setRendererSize(width: number, height: number): void;
    render(): void;
    destroy(): void;
    static createVisualizer(audioContext: AudioContext, canvas: HTMLCanvasElement, options?: VisualizerOptions): Visualizer;
  }

  export = Vissonance;
}

declare module 'vissonance/presets' {
  interface Preset {
    name: string;
    description?: string;
    class: any;
  }

  export function getPresets(): Record<string, Preset>;
  export const iris: Preset;
  export const barred: Preset;
  export const hillfog: Preset;
  export const tricentric: Preset;
  export const fracture: Preset;
  export const siphon: Preset;
  export const silk: Preset;
} 