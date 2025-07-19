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

interface AudioState {
  isPlaying: boolean;
  isLoading: boolean;
  currentBuffer: AudioBuffer | null;
  nextBuffer: AudioBuffer | null;
  elapsedTime: number;
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
  startPlaybackOnMount: () => void;
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

// Utility functions for localStorage operations
const saveVolumeSettings = (volume: number, muted: boolean) => {
  try {
    localStorage.setItem('go-radio-volume', volume.toString());
    localStorage.setItem('go-radio-muted', muted.toString());
  } catch (err) {
  }
};

const getStoredVolumeSettings = () => {
  try {
    const savedVolume = localStorage.getItem('go-radio-volume');
    const savedMuted = localStorage.getItem('go-radio-muted');
    
    const volume = savedVolume ? parseFloat(savedVolume) : 0.5;
    const muted = savedMuted ? JSON.parse(savedMuted) : false;
    
    // Validate and clamp values
    const validVolume = isNaN(volume) ? 0.5 : Math.max(0, Math.min(1, volume));
    const validMuted = typeof muted === 'boolean' ? muted : false;
    
    
    return { volume: validVolume, muted: validMuted };
  } catch (err) {
    return { volume: 0.5, muted: false };
  }
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
  
  // Audio state
  const [audioState, setAudioState] = useState<AudioState>({
    isPlaying: false,
    isLoading: false,
    currentBuffer: null,
    nextBuffer: null,
    elapsedTime: 0,
  });
  
  const [isAudioContextReady, setIsAudioContextReady] = useState(false);
  const [isQueueLoading] = useState(false);
  const [isUserInteracted, setIsUserInteracted] = useState(false);
  
  // Load saved volume settings
  const savedSettings = getStoredVolumeSettings();
  const [isMuted, setIsMuted] = useState(savedSettings.muted);
  const [volume, setVolume] = useState(savedSettings.volume);
  
  // Audio refs
  const audioContextRef = useRef<AudioContext | null>(null);
  const gainNodeRef = useRef<GainNode | null>(null);
  const sourceNodeRef = useRef<AudioBufferSourceNode | null>(null);
  const currentSongStartTimeRef = useRef<number>(0);
  const isPlaybackStartingRef = useRef<boolean>(false);

  // Audio management functions
  const stopCurrentAudio = useCallback((fromStartPlayback = false) => {
    if (sourceNodeRef.current) {
      try {
        sourceNodeRef.current.stop();
        sourceNodeRef.current.disconnect();
      } catch (err) {
        // Audio might already be stopped/disconnected
      }
      sourceNodeRef.current = null;
    }
    // Only cancel ongoing playback start if this isn't called from within startPlayback
    if (!fromStartPlayback) {
      isPlaybackStartingRef.current = false;
    }
    setAudioState(prev => ({ ...prev, isPlaying: false }));
  }, []);

  const initAudioContext = useCallback(() => {
    if (!audioContextRef.current && isUserInteracted) {
      try {
        audioContextRef.current = new AudioContext();
        gainNodeRef.current = audioContextRef.current.createGain();
        gainNodeRef.current.connect(audioContextRef.current.destination);
        
        // Set initial volume from saved state
        const initialVolume = isMuted ? 0 : volume;
        gainNodeRef.current.gain.setValueAtTime(
          initialVolume,
          audioContextRef.current.currentTime
        );
        
        setIsAudioContextReady(true);
      } catch (err) {
        toast.error("Failed to initialize audio system");
      }
    }
  }, [isMuted, volume, isUserInteracted]);

  const calculateElapsedTime = useCallback((startTime: string, duration?: number) => {
    const now = new Date();
    const startTimeDate = new Date(startTime);
    const elapsed = (now.getTime() - startTimeDate.getTime()) / 1000;
    if (duration) {
      return Math.min(Math.max(0, elapsed), duration);
    }
    return Math.max(0, elapsed);
  }, []);

  const getCurrentSong = useCallback(() => {
    if (!queueInfo.Queue || queueInfo.Queue.length === 0) return null;
    return queueInfo.Queue[queueInfo.CurrentSongIndex];
  }, [queueInfo.Queue, queueInfo.CurrentSongIndex]);

  const startPlayback = useCallback(async (audioBuf: ArrayBuffer, elapsed: number = 0) => {
    if (!audioContextRef.current || !gainNodeRef.current) {
      return;
    }

    // Prevent multiple simultaneous playback attempts
    if (isPlaybackStartingRef.current) {
      return;
    }


    try {
      isPlaybackStartingRef.current = true;
      
      // Stop any current audio first
      stopCurrentAudio(true);
      
      // Small delay to ensure previous audio is fully stopped
      await new Promise(resolve => setTimeout(resolve, 10));

      setAudioState(prev => ({ ...prev, isLoading: true }));

      // Decode the audio data
      const arrayBufferCopy = audioBuf.slice(0);
      const audioBuffer = await audioContextRef.current.decodeAudioData(arrayBufferCopy);
      
      
      // Create and configure new source
      const newSource = audioContextRef.current.createBufferSource();
      newSource.buffer = audioBuffer;
      newSource.connect(gainNodeRef.current);
      
      // Set up source reference and start playback
      sourceNodeRef.current = newSource;
      currentSongStartTimeRef.current = audioContextRef.current.currentTime - elapsed;
      
      // Handle song end
      newSource.onended = () => {
        setAudioState(prev => ({ ...prev, isPlaying: false }));
        sourceNodeRef.current = null;
        isPlaybackStartingRef.current = false;
      };

      // Start playback at the correct elapsed time
      newSource.start(0, elapsed);
      
      setAudioState(prev => ({ 
        ...prev, 
        isPlaying: true, 
        isLoading: false,
        currentBuffer: audioBuffer
      }));

    } catch (err) {
      toast.error("Failed to start audio playback");
      setAudioState(prev => ({ ...prev, isLoading: false, isPlaying: false }));
    } finally {
      isPlaybackStartingRef.current = false;
    }
  }, [stopCurrentAudio]);

  const startPlaybackOnMount = useCallback(async () => {
    if (!currentSongFile || !isAudioContextReady) return;
    
    // Don't start if playback is already starting (e.g., from song change)
    if (isPlaybackStartingRef.current) {
      return;
    }

    const currentSong = getCurrentSong();
    if (!currentSong) return;

    const elapsed = calculateElapsedTime(queueInfo.StartTime, currentSong.duration);
    
    await startPlayback(currentSongFile, elapsed);
  }, [currentSongFile, isAudioContextReady, queueInfo.StartTime, getCurrentSong, calculateElapsedTime, startPlayback]);

  const fetchQueue = async () => {
    try {
      setQueueLoading(true);
      const response = await api.get("/queue");
      setQueueInfo(response.data);
      return response.data as QueueInfo;
    } catch (err) {
      toast.error("Failed to fetch queue");
    } finally {
      setQueueLoading(false);
    }
  };

  const fetchSongFile = async (youtubeId: string): Promise<ArrayBuffer | null> => {
    try {
      const response = await fetch(`/api/v1/playlists/${youtubeId}/file`);
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      const arrayBuffer = await response.arrayBuffer();
      return arrayBuffer;
    } catch (err) {
      toast.error("Failed to load audio file");
      return null;
    }
  };

  const preloadNextSong = useCallback(async (nextSongIndex: number, queue: Song[]) => {
    if (!queue || nextSongIndex >= queue.length) {
      setNextSongFile(null);
      return;
    }

    const nextSong = queue[nextSongIndex];
    if (!nextSong) return;

    try {
      setNextSongFileLoading(true);
      const nextAudioBuffer = await fetchSongFile(nextSong.youtube_id);
      setNextSongFile(nextAudioBuffer);
      setNextSongFileError(null);
    } catch (err) {
      setNextSongFileError(err);
    } finally {
      setNextSongFileLoading(false);
    }
  }, []);

  const handleSongChange = useCallback(async (payload: {
    queue: Song[];
    playlist: Playlist | null;
    remaining: number;
    start_time: string;
    current_song_index: number;
  }) => {

    try {
      // Stop current audio immediately
      stopCurrentAudio();

      // Update queue info first
      const newQueueInfo = {
        Queue: payload.queue,
        Playlist: payload.playlist,
        Remaining: payload.remaining,
        StartTime: payload.start_time,
        CurrentSongIndex: payload.current_song_index,
      };
      setQueueInfo(newQueueInfo);

      // Determine what audio to play
      let audioToPlay: ArrayBuffer | null = null;

      // If we have the next song preloaded and it matches what we need, use it
      const currentSong = payload.queue[payload.current_song_index];
      if (nextSongFile && currentSong) {
        audioToPlay = nextSongFile;
        setCurrentSongFile(nextSongFile);
        setNextSongFile(null); // Clear it since we're using it
      } else {
        // Fetch the current song
        setCurrentSongFileLoading(true);
        audioToPlay = await fetchSongFile(currentSong.youtube_id);
        setCurrentSongFile(audioToPlay);
        setCurrentSongFileLoading(false);
        setCurrentSongFileError(null);
      }

      // Start playback immediately (song changes start from beginning)
      if (audioToPlay && isAudioContextReady) {
        await startPlayback(audioToPlay, 0);
      } else {
      }

      // Preload the next song
      const nextSongIndex = payload.current_song_index + 1;
      preloadNextSong(nextSongIndex, payload.queue);

    } catch (err) { 
      toast.error("Failed to change song");
    }
  }, [stopCurrentAudio, nextSongFile, isAudioContextReady, startPlayback, preloadNextSong]);

  const handleUserInteraction = useCallback(() => {
    setIsUserInteracted(true);  
    initAudioContext();
  }, [initAudioContext]);

  const handleVolumeChange = useCallback((newVolume: number) => {
    if (!audioContextRef.current || !gainNodeRef.current) return;

    // Clamp volume between 0 and 1
    const clampedVolume = Math.max(0, Math.min(1, newVolume));

    gainNodeRef.current.gain.setValueAtTime(
      isMuted ? 0 : clampedVolume, 
      audioContextRef.current.currentTime
    );
    
    setVolume(clampedVolume);
    
    // Update mute state based on volume
    const newMutedState = clampedVolume === 0 ? true : (clampedVolume > 0 && isMuted ? false : isMuted);
    setIsMuted(newMutedState);
    
    // Save both volume and mute state
    saveVolumeSettings(clampedVolume, newMutedState);
  }, [isMuted]);

  const toggleMute = useCallback(() => {
    if (!audioContextRef.current || !gainNodeRef.current) return;
    
    const newMutedState = !isMuted;
    gainNodeRef.current.gain.setValueAtTime(
      newMutedState ? 0 : volume, 
      audioContextRef.current.currentTime
    );
    setIsMuted(newMutedState);
    
    // Save settings
    saveVolumeSettings(volume, newMutedState);
  }, [isMuted, volume]);

  // Log initial volume settings on mount
  useEffect(() => {
  }, []); // Only run once on mount

  // Initialize on mount
  useEffect(() => {
    const handleMount = async () => {
      try {
        
        const queueRes = await fetchQueue();
        if (!queueRes || !queueRes.Queue.length) return;

        // Load current song
        setCurrentSongFileLoading(true);
        const currentSongBuffer = await fetchSongFile(queueRes.Queue[queueRes.CurrentSongIndex].youtube_id);
        setCurrentSongFile(currentSongBuffer);
        setCurrentSongFileLoading(false);
        setCurrentSongFileError(null);

        // Preload next song
        const nextIndex = queueRes.CurrentSongIndex + 1;
        preloadNextSong(nextIndex, queueRes.Queue);

      } catch (err) {
        toast.error("Failed to initialize radio");
      }
    };

    handleMount();
  }, [preloadNextSong]);

  // User interaction listeners
  useEffect(() => {
    document.addEventListener("click", handleUserInteraction, { once: true });
    document.addEventListener("touchstart", handleUserInteraction, { once: true });
    
    return () => {
      document.removeEventListener("click", handleUserInteraction);
      document.removeEventListener("touchstart", handleUserInteraction);
    };
  }, [handleUserInteraction]);

  // WebSocket event handlers
  useWebSocketEvent('song_change', (data) => {
    handleSongChange({
      queue: data.queue,
      playlist: data.playlist,
      remaining: data.remaining,
      start_time: data.start_time,
      current_song_index: data.current_song_index,
    });
  }, [handleSongChange]);

  useWebSocketEvent('queue_update', (data) => {
    setQueueInfo({
      Queue: data.queue,
      Playlist: data.playlist,
      Remaining: data.remaining,
      StartTime: data.start_time,
      CurrentSongIndex: data.current_song_index,
    });
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      stopCurrentAudio();
      if (audioContextRef.current) {
        audioContextRef.current.close();
      }
    };
  }, [stopCurrentAudio]);

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
    isPlaying: audioState.isPlaying,
    isAudioLoading: audioState.isLoading,
    isAudioContextReady,
    isWebSocketConnected,
    isQueueLoading,
    isUserInteracted,
    isMuted,
    volume,
    audioContextRef,
    gainNodeRef,
    isReady: isAudioContextReady && isWebSocketConnected && queueInfo.Queue.length > 0,
    setVolume,
    setIsMuted,
    setIsUserInteracted,
    initAudioContext,
    startPlayback,
    startPlaybackOnMount,
    getCurrentSong,
    calculateElapsedTime,
    handleVolumeChange,
    toggleMute,
  };

  return (
    <RadioContext.Provider value={value}>{children}</RadioContext.Provider>
  );
};