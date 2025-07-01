import { createContext, useCallback, useContext, useEffect, useRef, useState } from "react";
import toast from "react-hot-toast";
import api from "../lib/axios";
import { useWebSocket, useWebSocketEvent } from "./WebSocketContext";

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

export interface Playlist {
  id: number;
  name: string;
  description: string;
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
  handleVolumeChange: (newVolume: number) => void;
  toggleMute: () => void;

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

export const RadioProvider: React.FC<{ children?: React.ReactNode }> = ({ children }) => {
  // Use centralized WebSocket connection
  const { isConnected: isWebSocketConnected } = useWebSocket();
  
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
  const [isQueueLoading] = useState(false);
  const [isUserInteracted, setIsUserInteracted] = useState(false);
  const [isMuted, setIsMuted] = useState(false);
  const [volume, setVolume] = useState(0.5);
  
  // Refs
  const audioContextRef = useRef<AudioContext | null>(null);
  const gainNodeRef = useRef<GainNode | null>(null);
  const sourceNodeRef = useRef<AudioBufferSourceNode | null>(null);

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
      toast.error("Failed getting song file");
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
        console.log("stopping current song")
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
  }, [audioContextRef, gainNodeRef, sourceNodeRef, nextSongFile, fetchSongFile, startPlayback]);


  const handleUserInteraction = useCallback(() => {
    setIsUserInteracted(true);  
    initAudioContext();
  }, [])

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

  const handleVolumeChange = (newVolume: number) => {
    if (!audioContextRef.current || !gainNodeRef.current) return;

    console.log("setting volume", newVolume)

    gainNodeRef.current.gain.setValueAtTime(newVolume, audioContextRef.current.currentTime);
    setVolume(newVolume);
    if (newVolume === 0) {
      setIsMuted(true);
    } else if (isMuted) {
      setIsMuted(false);
    }
  };

  const handleMute = () => {
    if (!audioContextRef.current || !gainNodeRef.current) return;
    gainNodeRef.current.gain.setValueAtTime(0, audioContextRef.current.currentTime);
    setIsMuted(true);
  }

  const handleUnmute = () => {
    if (!audioContextRef.current || !gainNodeRef.current) return;
    gainNodeRef.current.gain.setValueAtTime(volume, audioContextRef.current.currentTime);
    setIsMuted(false);
  }

  const toggleMute = () => {
    if (isMuted) {
      handleUnmute();
    } else {
      handleMute();
    }
  }

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
        toast.error("/handlFailed to fetch queue");
      }
    };

    handleMount();
  }, [audioContextRef, gainNodeRef]);

  // Subscribe to WebSocket events using the centralized event bus
  useWebSocketEvent('song_change', (data) => {
    handleSongChange({
      queue: data.queue,
      playlist: data.playlist,
      remaining: data.remaining,
      start_time: data.start_time,
      current_song_index: data.current_song_index,
    });
  }, []);

  useWebSocketEvent('queue_update', (data) => {
    setQueueInfo({
      Queue: data.queue,
      Playlist: data.playlist,
      Remaining: data.remaining,
      StartTime: data.start_time,
      CurrentSongIndex: data.current_song_index,
    });
  }, []);


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
    handleVolumeChange,
    toggleMute,
  }

  return (
    <RadioContext.Provider value={value}>{children}</RadioContext.Provider>
  );
};