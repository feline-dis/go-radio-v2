import { createContext, useCallback, useContext, useEffect, useRef, useState } from "react";
import toast from "react-hot-toast";
import api from "../lib/axios";

const RadioContext = createContext<RadioContextType | undefined>(undefined);

export interface SongChangeEvent {
  Queue: Song[];
  Playlist: {
    id: number;
    name: string;
    description: string;
  } | null;
  Remaining: number;
  StartTime: string;
  CurrentSongIndex: number;
}

export interface Song {
  id: number;
  youtube_id: string;
  title: string;
  description: string;
  duration: number;
  position: number;
}

export interface QueueInfo {
  Queue: Song[];
  Playlist: {
    id: number;
    name: string;
    description: string;
  } | null;
  Remaining: number;
  StartTime: string;
  CurrentSongIndex: number;
}

interface RadioContextType {
  queueInfo: QueueInfo;
  queueError: unknown;
  queueLoading: boolean;
  currentSongFile: ArrayBuffer | null;
  currentSongFileLoading: boolean;
  currentSongFileError: unknown;
  nextSongFile: ArrayBuffer | null;
  nextSongFileLoading: boolean;
  nextSongFileError: unknown;
  isPlaying: boolean;
  isAudioLoading: boolean;
  isAudioContextReady: boolean;
  isWebSocketConnected: boolean;
  isQueueLoading: boolean;
  isUserInteracted: boolean;
  isMuted: boolean;
  volume: number;
  isReady: boolean;
  setVolume: (volume: number) => void;
  setIsMuted: (muted: boolean) => void;
  setIsUserInteracted: (interacted: boolean) => void;
  initAudioContext: () => void;
  startPlayback: (audioBuf: ArrayBuffer, elapsed: number) => void;
  getCurrentSong: () => Song | null;
  calculateElapsedTime: (startTime: string, duration?: number) => number;


  // Audio refs for visualizer
  audioContextRef: React.MutableRefObject<AudioContext | null>;
  gainNodeRef: React.MutableRefObject<GainNode | null>;
}

export const useRadio = () => {
  const context = useContext(RadioContext);
  if (context === undefined) {
    throw new Error("useRadio must be used within a RadioProvider :p");
  }
  return context;
};

export const RadioProvider: React.FC<{ children?: React.ReactNode, wsUrl: string }> = ({ children, wsUrl }) => {
  const [queueInfo, setQueueInfo] = useState<QueueInfo>({
    Queue: [],
    Playlist: null,
    Remaining: 0,
    StartTime: "",
    CurrentSongIndex: 0,
  });
  const [queueError] = useState<unknown>(null);
  const [queueLoading, setQueueLoading] = useState(false);
  const [currentSongFile, setCurrentSongFile] = useState<ArrayBuffer | null>(null);
  const [currentSongFileLoading, setCurrentSongFileLoading] = useState(false);
  const [currentSongFileError, setCurrentSongFileError] = useState<unknown>(null);
  const [nextSongFile, setNextSongFile] = useState<ArrayBuffer | null>(null);
  const [nextSongFileLoading, setNextSongFileLoading] = useState(false);
  const [nextSongFileError, setNextSongFileError] = useState<unknown>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [isAudioLoading, setIsAudioLoading] = useState(false);
  const [isAudioContextReady, setIsAudioContextReady] = useState(false);
  const [isWebSocketConnected, setIsWebSocketConnected] = useState(false);
  const [isQueueLoading] = useState(false);
  const [isUserInteracted, setIsUserInteracted] = useState(false);
  const [isMuted, setIsMuted] = useState(false);
  const [volume, setVolume] = useState(0.5);
  // Refs
  const audioContextRef = useRef<AudioContext | null>(null);
  const gainNodeRef = useRef<GainNode | null>(null);
  const sourceNodeRef = useRef<AudioBufferSourceNode | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

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

  const startPlayback = useCallback(async (audioBuf: ArrayBuffer, elapsed: number) => {
    if (!audioContextRef.current || !gainNodeRef.current) return;
    const arrayBufferCopy = audioBuf.slice(0);
    const audioBuffer = await audioContextRef.current.decodeAudioData(arrayBufferCopy);
    sourceNodeRef.current = audioContextRef.current.createBufferSource();
    sourceNodeRef.current.buffer = audioBuffer;
    sourceNodeRef.current.connect(gainNodeRef.current);
    sourceNodeRef.current.start(0, elapsed);
    setIsPlaying(true);
    setIsAudioLoading(false);
  }, [audioContextRef, gainNodeRef, sourceNodeRef]);

  const fetchQueue = async () => {
    try {
      setQueueLoading(true);
      const response = await api.get("/queue");
      setQueueInfo(response.data);
      return response.data as QueueInfo;
    } catch (error) {
      toast.error("Failed to fetch queue");
    } finally {
      setQueueLoading(false);
    }
  };

  const fetchSongFile = async (youtubeId: string) => {
    try {
      const response = await fetch(`/api/v1/playlists/${youtubeId}/file`);
      const arrayBuffer = await response.arrayBuffer();

      return arrayBuffer;
    } catch (error) {
      toast.error("Failed to fetch song file");
      return null;
    }
  }

  const handleSongChange = useCallback(async (payload: any) => {
    console.log("handleSongChange", payload);

    try {
      if (!audioContextRef.current || !gainNodeRef.current) return;

      console.log("changing song", payload.current_song_index)

      let toPlay: ArrayBuffer | null = null;

      if (!nextSongFile) {
        toPlay = await fetchSongFile(payload.queue[payload.current_song_index].youtube_id);
      } else {
        toPlay = nextSongFile;
      }

      console.log("toPlay", toPlay);

      if (sourceNodeRef.current) {
        sourceNodeRef.current.stop();
        sourceNodeRef.current.disconnect();
        sourceNodeRef.current = null;
      }

      console.log("starting playback")
      startPlayback(toPlay as ArrayBuffer, 0);
      setCurrentSongFile(toPlay as ArrayBuffer);
      setCurrentSongFileLoading(false);
      setCurrentSongFileError(null);

      console.log("setting next song file")
      if (
        payload.current_song_index + 1 < payload.queue.length
      ) {
        const nextNextSongFile = await fetchSongFile(payload.queue[payload.current_song_index + 1].youtube_id);
        setNextSongFile(nextNextSongFile);
        setNextSongFileLoading(false);
        setNextSongFileError(null);
      } else {
        setNextSongFile(null);
        setNextSongFileLoading(false);
        setNextSongFileError(null);
      }

      setQueueInfo({
        Queue: payload.queue,
        Playlist: payload.playlist,
        Remaining: payload.remaining,
        StartTime: payload.start_time,
        CurrentSongIndex: payload.current_song_index,
      });
    } catch (error) { 
      console.error("Error changing song:", error);
      toast.error("Failed to change song");
    }
  }, []);


  const handleUserInteraction = useCallback(() => {
    setIsUserInteracted(true);  
    initAudioContext();
  }, [])

  useEffect(() => {
    document.addEventListener("click", handleUserInteraction);
    document.addEventListener("touchstart", handleUserInteraction);
    return () => {
      document.removeEventListener("click", handleUserInteraction);
      document.removeEventListener("touchstart", handleUserInteraction);
    }
  }, [handleUserInteraction])

  useEffect(() => {
    const handleMount = async () => {
      try {
        console.log("fetching queue")

        const queueRes = await fetchQueue();

        if (!queueRes) return;

        const songFileRes = await fetchSongFile(queueRes.Queue[queueRes.CurrentSongIndex].youtube_id);
        setCurrentSongFile(songFileRes)
        setCurrentSongFileLoading(false);
        setCurrentSongFileError(null);

        const nextSongFileRes = await fetchSongFile(queueRes.Queue[queueRes.CurrentSongIndex + 1].youtube_id);
        setNextSongFile(nextSongFileRes);
        setNextSongFileLoading(false);
        setNextSongFileError(null);
      } catch (error) {
        console.error("Error fetching queue:", error);
        toast.error("Failed to fetch queue");
      }
    };

    handleMount();
  }, [audioContextRef, gainNodeRef]);

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

  const calculateElapsedTime = (startTime: string, duration?: number) => {
    const now = new Date();
    const startTimeDate = new Date(startTime);
    const elapsed = (now.getTime() - startTimeDate.getTime()) / 1000;
    if (duration) {
      return Math.min(elapsed, duration);
    }
    return Math.max(0, elapsed);
  }

  const getCurrentSong = () => {
    if (!queueInfo.Queue || queueInfo.Queue.length === 0) return null;
    return queueInfo.Queue[queueInfo.CurrentSongIndex];
  }

  const value: RadioContextType = {
    queueInfo,
    queueError,
    queueLoading,
    currentSongFile,
    currentSongFileLoading,
    currentSongFileError,
    nextSongFile,
    nextSongFileLoading,
    nextSongFileError,
    isPlaying,
    isAudioLoading,
    isAudioContextReady,
    isWebSocketConnected,
    isQueueLoading,
    isUserInteracted,
    isMuted,
    volume,
    audioContextRef,
    gainNodeRef,
    isReady: isAudioContextReady && isWebSocketConnected,
    setVolume,
    setIsMuted,
    setIsUserInteracted,
    initAudioContext,
    startPlayback,
    getCurrentSong,
    calculateElapsedTime,
  }

  return (
    <RadioContext.Provider value={value}>{children}</RadioContext.Provider>
  );
};