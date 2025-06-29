import { SpeakerWaveIcon, SpeakerXMarkIcon } from "@heroicons/react/24/solid";
import { useRadio } from "../contexts/RadioContext";
import { useEffect, useRef, useState } from "react";
import { AnimatedEmotes } from "./AnimatedEmotes";

interface Song {
  youtube_id: string;
  title: string;
  artist: string;
  album: string;
  duration: number;
  s3_key: string;
}

interface QueueInfo {
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

// Helper functions to derive current and next songs from queue and index
const getCurrentSong = (queueInfo: QueueInfo | null): Song | null => {
  if (!queueInfo || !queueInfo.Queue || queueInfo.Queue.length === 0) {
    return null;
  }
  
  const currentIndex = queueInfo.CurrentSongIndex;
  if (currentIndex < 0 || currentIndex >= queueInfo.Queue.length) {
    return null;
  }
  
  return queueInfo.Queue[currentIndex];
};



// Separate ProgressBar component that handles its own updates
const ProgressBar = ({
  currentSong,
  elapsed,
}: {
  currentSong: Song | null;
  elapsed: number;
}) => {
  const [localElapsed, setLocalElapsed] = useState(elapsed);
  const lastUpdateRef = useRef(Date.now());
  const rafRef = useRef<number | null>(null);

  useEffect(() => {
    setLocalElapsed(elapsed);
  }, [elapsed]);

  useEffect(() => {
    const updateElapsed = () => {
      const now = Date.now();
      if (now - lastUpdateRef.current > 100) {
        lastUpdateRef.current = now;
        setLocalElapsed((prev) => prev + 0.1);
      }
      rafRef.current = requestAnimationFrame(updateElapsed);
    };

    rafRef.current = requestAnimationFrame(updateElapsed);

    return () => {
      if (rafRef.current) {
        cancelAnimationFrame(rafRef.current);
      }
    };
  }, []);

  const progress = (localElapsed / (currentSong?.duration || 0)) * 100;
  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };

  return (
    <div className="mb-2">
      <div className="relative h-1 bg-black border border-gray-700 rounded-none overflow-hidden">
        <div
          className="absolute top-0 left-0 h-full bg-white transition-all duration-1000"
          style={{ width: `${progress}%` }}
        />
      </div>
      <div className="flex justify-between text-xs text-gray-500 mt-2 font-mono">
        <span>{formatTime(localElapsed)}</span>
        <span>{currentSong ? formatTime(currentSong.duration) : "0:00"}</span>
      </div>
    </div>
  );
};

export const RadioPlayer = () => {
  const {
    queueInfo,
    elapsed,
    volume,
    isMuted,
    isAudioLoading,
    isAudioContextReady,
    isPlaying,
    setVolume,
    setIsMuted,
  } = useRadio();

  const currentSong = getCurrentSong(queueInfo);

  const handleVolumeChange = (newVolume: number) => {
    setVolume(newVolume);
    if (newVolume === 0) {
      setIsMuted(true);
    } else if (isMuted) {
      setIsMuted(false);
    }
  };

  const toggleMute = () => {
    setIsMuted(!isMuted);
  };

  return (
    <>
      <AnimatedEmotes />
      <div className="max-w-4xl mx-auto p-4 bg-black border border-gray-800 shadow-2xl">
        {/* Header */}
        <div className="text-center mb-6">
          {!isAudioContextReady && (
            <p className="text-xs text-yellow-500 mb-3 font-mono">
              [INITIALIZING AUDIO SYSTEM...]
            </p>
          )}
        </div>

        {/* Current Song Info */}
        <div className="text-center mb-6">
          {currentSong ? (
            <>
              <p className="text-lg text-white font-mono truncate">
                {currentSong.title}
              </p>
              {isAudioLoading && (
                <p className="text-xs text-white font-mono">
                  [LOADING AUDIO...]
                </p>
              )}
              {isPlaying && (
                <p className="text-xs text-green-500 font-mono">[PLAYING]</p>
              )}
            </>
          ) : (
            <div className="bg-gray-900 border border-gray-700 p-4">
              <p className="text-lg text-gray-500 font-mono">
                [NO TRACK ACTIVE]
              </p>
            </div>
          )}
        </div>

        {/* Progress Bar */}
        <ProgressBar
          currentSong={currentSong}
          elapsed={elapsed}
        />

        {/* Volume Controls */}
        <div className="flex gap-4 mb-2">
          <button
            onClick={toggleMute}
            className="p-1 text-gray-600 hover:text-white transition-colors border border-gray-700 hover:border-white rounded-sm"
            aria-label={isMuted ? "Unmute" : "Mute"}
            disabled={!isAudioContextReady}
          >
            {isMuted || volume === 0 ? (
              <SpeakerXMarkIcon className="h-4 w-4" />
            ) : (
              <SpeakerWaveIcon className="h-4 w-4" />
            )}
          </button>

          <div className="flex flex-1 items-center gap-3 w-40">
            <input
              type="range"
              min="0"
              max="1"
              step="0.01"
              value={isMuted ? 0 : volume}
              onChange={(e) => handleVolumeChange(parseFloat(e.target.value))}
              className="w-full h-1 bg-black border border-gray-700 appearance-none cursor-pointer slider"
              style={{
                background:
                  "linear-gradient(to right, #ffffff 0%, #ffffff " +
                  (isMuted ? 0 : volume * 100) +
                  "%, #333333 " +
                  (isMuted ? 0 : volume * 100) +
                  "%, #333333 100%)",
              }}
              disabled={!isAudioContextReady}
            />
          </div>
        </div>



        {/* Queue Info */}
        {queueInfo?.Queue && queueInfo.Queue.length > 0 && (
          <div className="mt-6">
            <h3 className="text-sm text-gray-500 font-mono mb-2">[QUEUE]</h3>
            <div className="space-y-1">
              {/* Show current song */}
              {currentSong && (
                <div className="text-xs font-mono text-white">
                  ▶ {currentSong.title}
                </div>
              )}
              {/* Show next 2 upcoming songs */}
              {queueInfo.Queue.slice(queueInfo.CurrentSongIndex + 1, queueInfo.CurrentSongIndex + 3).map((song, index) => (
                <div
                  key={song.youtube_id}
                  className="text-xs font-mono text-gray-500"
                >
                  {index + 2}. {song.title}
                </div>
              ))}
              {/* If we're near the end of the queue, show songs from the beginning */}
              {queueInfo.CurrentSongIndex + 3 > queueInfo.Queue.length && 
                queueInfo.Queue.slice(0, Math.min(2, queueInfo.Queue.length - 1)).map((song, index) => (
                  <div
                    key={`loop-${song.youtube_id}`}
                    className="text-xs font-mono text-gray-500"
                  >
                    {queueInfo.CurrentSongIndex + 2 + index}. {song.title}
                  </div>
                ))
              }
            </div>
          </div>
        )}
      </div>
    </>
  );
};
