import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useRef,
  useCallback,
} from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import api from "../lib/axios";
import { toast } from "react-hot-toast";

interface Song {
  youtube_id: string;
  title: string;
  artist: string;
  album: string;
  duration: number;
  s3_key: string;
}

interface QueueInfo {
  CurrentSong: Song | null;
  NextSong: Song | null;
  Queue: Song[];
  Playlist: {
    id: number;
    name: string;
    description: string;
  } | null;
  Remaining: number;
  StartTime: string;
}

interface RadioContextType {
  // State
  elapsed: number;
  isPaused: boolean;
  volume: number;
  isMuted: boolean;
  isAudioLoading: boolean;
  isAudioContextReady: boolean;
  isWebSocketConnected: boolean;
  isQueueLoading: boolean;
  isCurrentSongFileLoading: boolean;
  isUserInteracted: boolean;
  isPlaying: boolean;
  isVisualizerEnabled: boolean;
  currentVisualizerPreset: string;

  // Actions
  setElapsed: (elapsed: number) => void;
  setIsPaused: (paused: boolean) => void;
  setVolume: (volume: number) => void;
  setIsMuted: (muted: boolean) => void;
  setIsUserInteracted: (interacted: boolean) => void;
  setIsVisualizerEnabled: (enabled: boolean) => void;
  setCurrentVisualizerPreset: (preset: string) => void;

  // Data
  queueInfo: QueueInfo | null;
  queueError: unknown;
  currentSongFile: ArrayBuffer | null;

  // Functions
  startPlayback: () => void;
  pausePlayback: () => void;
  seekTo: (position: number) => void;
  initAudioContext: () => void;
  cleanup: () => void;
  refetchQueue: () => void;

  // Computed
  isReady: boolean;

  // Audio refs for visualizer
  audioContextRef: React.MutableRefObject<AudioContext | null>;
  gainNodeRef: React.MutableRefObject<GainNode | null>;
}

const RadioContext = createContext<RadioContextType | undefined>(undefined);

export const useRadio = () => {
  const context = useContext(RadioContext);
  if (context === undefined) {
    throw new Error("useRadio must be used within a RadioProvider");
  }
  return context;
};

interface RadioProviderProps {
  children: React.ReactNode;
  wsUrl: string;
}

export const RadioProvider: React.FC<RadioProviderProps> = ({
  children,
  wsUrl,
}) => {
  // State
  const [elapsed, setElapsed] = useState(0);
  const [isPaused, setIsPaused] = useState(false);
  const [volume, setVolume] = useState(0.7);
  const [isMuted, setIsMuted] = useState(false);
  const [isAudioLoading, setIsAudioLoading] = useState(false);
  const [isAudioContextReady, setIsAudioContextReady] = useState(false);
  const [isWebSocketConnected, setIsWebSocketConnected] = useState(false);
  const [isUserInteracted, setIsUserInteracted] = useState(false);
  const [isPlaying, setIsPlaying] = useState(false);
  const [isVisualizerEnabled, setIsVisualizerEnabled] = useState(false);
  const [currentVisualizerPreset, setCurrentVisualizerPreset] = useState("");

  // Refs
  const audioContextRef = useRef<AudioContext | null>(null);
  const sourceNodeRef = useRef<AudioBufferSourceNode | null>(null);
  const gainNodeRef = useRef<GainNode | null>(null);
  const startTimeRef = useRef<number>(0);
  const wsRef = useRef<WebSocket | null>(null);
  const syncIntervalRef = useRef<number | null>(null);
  const isStartingPlaybackRef = useRef<boolean>(false);

  const [queueInfo, setQueueInfo] = useState<QueueInfo | null>(null);
  const [currentSongFile, setCurrentSongFile] = useState<ArrayBuffer | null>(
    null
  );

  const queryClient = useQueryClient();

  // Fetch queue information
  const {
    error: queueError,
    isLoading: isQueueLoading,
    refetch: refetchQueue,
  } = useQuery<QueueInfo>({
    queryKey: ["queue"],
    queryFn: async () => {
      const response = await api.get("/queue");
      const data = response.data;

      setQueueInfo(data);
      return {
        CurrentSong: data.CurrentSong ?? null,
        NextSong: data.NextSong ?? null,
        Queue: data.Queue ?? [],
        Playlist: data.Playlist ?? null,
        Remaining: data.Remaining ?? 0,
        StartTime: data.StartTime ?? "",
      };
    },
    refetchOnWindowFocus: false,
    refetchInterval: 5000, // Refetch every 5 seconds to stay in sync
  });

  // Load current song file
  const { isLoading: isCurrentSongFileLoading } = useQuery({
    queryKey: ["currentSongFile", queueInfo?.CurrentSong?.youtube_id],
    queryFn: async () => {
      if (!queueInfo?.CurrentSong?.youtube_id) return null;

      const response = await fetch(
        `/api/v1/playlists/${queueInfo.CurrentSong.youtube_id}/file`
      );
      const arrayBuffer = await response.arrayBuffer();
      setCurrentSongFile(arrayBuffer);
      return arrayBuffer;
    },
    enabled: !!queueInfo?.CurrentSong?.youtube_id,
    staleTime: 0, // Always consider data stale to prevent caching issues
    gcTime: 0, // Don't cache to prevent detached buffer issues
    refetchOnMount: true,
    refetchOnWindowFocus: false,
  });

  // Initialize audio context on user interaction
  const initAudioContext = useCallback(() => {
    if (!audioContextRef.current) {
      try {
        audioContextRef.current = new AudioContext();
        gainNodeRef.current = audioContextRef.current.createGain();
        gainNodeRef.current.connect(audioContextRef.current.destination);
        gainNodeRef.current.gain.setValueAtTime(
          isMuted ? 0 : volume,
          audioContextRef.current.currentTime
        );
        setIsAudioContextReady(true);
        console.log("AudioContext initialized successfully");
      } catch (error) {
        console.error("Failed to initialize AudioContext:", error);
        toast.error("Failed to initialize audio system");
      }
    }
  }, [isMuted, volume]);

  // Calculate elapsed time based on server start time
  const calculateElapsedTime = useCallback(() => {
    if (!queueInfo?.StartTime) return 0;

    const now = new Date();
    const startTime = new Date(queueInfo.StartTime);
    const elapsed = (now.getTime() - startTime.getTime()) / 1000;

    // If song has duration, cap elapsed time
    if (queueInfo.CurrentSong?.duration) {
      return Math.min(elapsed, queueInfo.CurrentSong.duration);
    }

    return Math.max(0, elapsed);
  }, [queueInfo?.StartTime, queueInfo?.CurrentSong?.duration]);

  // Handle song changes from WebSocket
  const handleSongChange = useCallback(
    async (payload: SongChangeEvent) => {
      console.log("🎵 Received song change event:", payload);

      // Stop current playback immediately
      if (sourceNodeRef.current) {
        try {
          sourceNodeRef.current.stop();
          sourceNodeRef.current.disconnect();
          sourceNodeRef.current = null;
        } catch (error) {
          console.error("Error stopping current playback:", error);
        }
      }

      // Reset playback state and flags
      setIsPlaying(false);
      setIsPaused(false);
      setElapsed(0);
      isStartingPlaybackRef.current = false;

      // Update queue info
      setQueueInfo(payload);

      // Invalidate queries to refetch data
      queryClient.invalidateQueries({ queryKey: ["queue"] });
      queryClient.invalidateQueries({ queryKey: ["currentSongFile"] });

      // Refetch queue to get updated information
      await refetchQueue();
    },
    [queryClient, refetchQueue]
  );

  // Start playback synchronized with server
  const startPlayback = useCallback(async () => {
    if (!audioContextRef.current || !gainNodeRef.current || !currentSongFile) {
      console.log("Cannot start playback - missing audio context or song file");
      return;
    }

    // Prevent concurrent playback attempts
    if (isStartingPlaybackRef.current) {
      console.log("Playback already starting, skipping...");
      return;
    }

    try {
      isStartingPlaybackRef.current = true;
      setIsAudioLoading(true);

      // Stop any existing playback first
      if (sourceNodeRef.current) {
        try {
          sourceNodeRef.current.stop();
          sourceNodeRef.current.disconnect();
          sourceNodeRef.current = null;
        } catch (error) {
          console.error("Error stopping existing playback:", error);
        }
      }

      // Create a copy of the ArrayBuffer to avoid detached buffer issues
      const audioBufferCopy = currentSongFile.slice(0);

      // Decode audio data
      const audioBuffer = await audioContextRef.current.decodeAudioData(
        audioBufferCopy
      );

      // Calculate start position based on server time
      const startPosition = calculateElapsedTime();

      // Create new source node
      sourceNodeRef.current = audioContextRef.current.createBufferSource();
      sourceNodeRef.current.buffer = audioBuffer;
      sourceNodeRef.current.connect(gainNodeRef.current);

      // Start playback from calculated position
      sourceNodeRef.current.start(0, startPosition);
      startTimeRef.current =
        audioContextRef.current.currentTime - startPosition;

      setIsPlaying(true);
      setIsPaused(false);
      setElapsed(startPosition);

      console.log(`Started playback at position: ${startPosition}s`);
    } catch (error) {
      console.error("Failed to start playback:", error);
      toast.error("Failed to start playback");
    } finally {
      setIsAudioLoading(false);
      isStartingPlaybackRef.current = false;
    }
  }, [currentSongFile, calculateElapsedTime]);

  // Pause playback
  const pausePlayback = useCallback(() => {
    if (sourceNodeRef.current) {
      sourceNodeRef.current.stop();
      sourceNodeRef.current.disconnect();
      sourceNodeRef.current = null;
      setIsPlaying(false);
      setIsPaused(true);
    }
  }, []);

  // Seek to position
  const seekTo = useCallback(
    (position: number) => {
      const newPosition = Math.max(0, position);
      setElapsed(newPosition);

      if (isPlaying) {
        // Restart playback at new position
        startPlayback();
      }
    },
    [isPlaying, startPlayback]
  );

  // WebSocket connection
  useEffect(() => {
    const connectWebSocket = () => {
      console.log("Connecting to radio server at:", wsUrl);
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log("✅ Connected to radio server successfully");
        setIsWebSocketConnected(true);
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);

          switch (data.type) {
            case "song_change":
              handleSongChange(data.payload);
              break;
            case "pong":
              console.log("🏓 Received pong from server");
              break;
            default:
              console.log("❓ Unknown WebSocket message type:", data.type);
          }
        } catch (error) {
          console.error("❌ Failed to parse WebSocket message:", error);
        }
      };

      ws.onerror = (error) => {
        console.error("❌ WebSocket error:", error);
        setIsWebSocketConnected(false);
        toast.error("Lost connection to radio server");
      };

      ws.onclose = (event) => {
        console.log("🔌 Disconnected from radio server. Code:", event.code);
        setIsWebSocketConnected(false);
        // Attempt to reconnect after 5 seconds
        setTimeout(connectWebSocket, 5000);
      };
    };

    connectWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [wsUrl, handleSongChange]);

  // Sync elapsed time with server
  useEffect(() => {
    if (isPlaying && queueInfo?.StartTime) {
      syncIntervalRef.current = setInterval(() => {
        const serverElapsed = calculateElapsedTime();
        setElapsed(serverElapsed);
      }, 100); // Update every 100ms for smooth progress
    } else {
      if (syncIntervalRef.current) {
        clearInterval(syncIntervalRef.current);
        syncIntervalRef.current = null;
      }
    }

    return () => {
      if (syncIntervalRef.current) {
        clearInterval(syncIntervalRef.current);
        syncIntervalRef.current = null;
      }
    };
  }, [isPlaying, queueInfo?.StartTime, calculateElapsedTime]);

  // Auto-start playback when new song file is loaded (after song change)
  useEffect(() => {
    if (
      currentSongFile &&
      isAudioContextReady &&
      !isPlaying &&
      !isPaused &&
      !isStartingPlaybackRef.current
    ) {
      console.log("🎵 Auto-starting playback for new song file");
      // Use setTimeout to ensure state updates have settled
      const timer = setTimeout(() => {
        startPlayback();
      }, 100);

      return () => clearTimeout(timer);
    }
  }, [
    currentSongFile,
    isAudioContextReady,
    isPlaying,
    isPaused,
    startPlayback,
  ]);

  // Handle volume changes
  useEffect(() => {
    if (gainNodeRef.current && audioContextRef.current) {
      const targetVolume = isMuted ? 0 : volume;
      gainNodeRef.current.gain.setTargetAtTime(
        targetVolume,
        audioContextRef.current.currentTime,
        0.1
      );
    }
  }, [volume, isMuted]);

  // Show error toast if queue query fails
  useEffect(() => {
    if (queueError) {
      toast.error("Failed to fetch queue information");
    }
  }, [queueError]);

  // Cleanup function
  const cleanup = useCallback(() => {
    if (sourceNodeRef.current) {
      try {
        sourceNodeRef.current.stop();
        sourceNodeRef.current.disconnect();
        sourceNodeRef.current = null;
      } catch (error) {
        console.error("Error stopping source node:", error);
      }
    }

    if (gainNodeRef.current) {
      try {
        gainNodeRef.current.disconnect();
        gainNodeRef.current = null;
      } catch (error) {
        console.error("Error disconnecting gain node:", error);
      }
    }

    if (audioContextRef.current) {
      try {
        audioContextRef.current.close();
        audioContextRef.current = null;
      } catch (error) {
        console.error("Error closing AudioContext:", error);
      }
    }

    if (wsRef.current) {
      try {
        wsRef.current.close();
        wsRef.current = null;
      } catch (error) {
        console.error("Error closing WebSocket:", error);
      }
    }

    if (syncIntervalRef.current) {
      clearInterval(syncIntervalRef.current);
      syncIntervalRef.current = null;
    }

    // Reset flags
    isStartingPlaybackRef.current = false;

    // Reset state
    setIsAudioContextReady(false);
    setIsWebSocketConnected(false);
    setIsAudioLoading(false);
    setIsPlaying(false);
    setIsPaused(false);
    setElapsed(0);
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      cleanup();
    };
  }, [cleanup]);

  // Computed property for readiness
  const isReady =
    isUserInteracted &&
    !isQueueLoading &&
    !isCurrentSongFileLoading &&
    isWebSocketConnected &&
    isAudioContextReady;

  const value: RadioContextType = {
    // State
    elapsed,
    isPaused,
    volume,
    isMuted,
    isAudioLoading,
    isAudioContextReady,
    isWebSocketConnected,
    isQueueLoading,
    isCurrentSongFileLoading,
    isUserInteracted,
    isPlaying,
    isVisualizerEnabled,
    currentVisualizerPreset,

    // Actions
    setElapsed,
    setIsPaused,
    setVolume,
    setIsMuted,
    setIsUserInteracted,
    setIsVisualizerEnabled,
    setCurrentVisualizerPreset,

    // Data
    queueInfo,
    queueError,
    currentSongFile,

    // Functions
    startPlayback,
    pausePlayback,
    seekTo,
    initAudioContext,
    cleanup,
    refetchQueue,

    // Computed
    isReady,

    // Audio refs for visualizer
    audioContextRef,
    gainNodeRef,
  };

  return (
    <RadioContext.Provider value={value}>{children}</RadioContext.Provider>
  );
};

interface SongChangeEvent {
  CurrentSong: Song | null;
  NextSong: Song | null;
  Queue: Song[];
  Playlist: {
    id: number;
    name: string;
    description: string;
  } | null;
  Remaining: number;
  StartTime: string;
}
