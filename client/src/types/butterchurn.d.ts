declare module "butterchurn" {
  export interface VisualizerOptions {
    width: number;
    height: number;
  }

  export interface Visualizer {
    connectAudio(audioNode: AudioNode): void;
    loadPreset(preset: unknown, blendTime: number): void;
    render(): void;
    setRendererSize(width: number, height: number): void;
    disconnect(): void;
  }

  export function createVisualizer(
    audioContext: AudioContext,
    canvas: HTMLCanvasElement,
    options: VisualizerOptions
  ): Visualizer;

  export default {
    createVisualizer,
  };
}

declare module "butterchurn-presets" {
  export function getPresets(): Record<string, unknown>;

  export default {
    getPresets,
  };
}
