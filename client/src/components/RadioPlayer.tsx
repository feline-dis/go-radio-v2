import { SpeakerWaveIcon, SpeakerXMarkIcon } from "@heroicons/react/24/solid";
import { ChevronUpIcon, ChevronDownIcon } from "@heroicons/react/24/outline";
import { useRadio } from "../contexts/RadioContext";
import { useEffect, useRef, useState } from "react";
import { VissonanceVisualizer, VissonancePresetSelector } from "./VissonanceVisualizer";
import { VisualizerToggle } from "./VisualizerToggle";
import { ReactionBar } from "./ReactionBar";

interface Song {
  id: number;
  youtube_id: string;
  title: string;
  description: string;
  duration: number;
  position: number;
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



// Separate ProgressBar component that handles its own updates
const ProgressBar = ({
  currentSong,
  queueInfo,
}: {
  currentSong: Song | null;
  queueInfo: QueueInfo | null;
}) => {
  const { calculateElapsedTime } = useRadio();
  const [localElapsed, setLocalElapsed] = useState(0);
  const lastUpdateRef = useRef(Date.now());
  const rafRef = useRef<number | null>(null);

  // Calculate initial elapsed time
  useEffect(() => {
    if (queueInfo && currentSong) {
      const elapsed = calculateElapsedTime(queueInfo.StartTime, currentSong.duration);
      setLocalElapsed(elapsed);
    }
  }, [queueInfo, currentSong]);

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
      <div className="flex justify-between text-xs text-gray-500 mt-1 sm:mt-2 font-mono">
        <span>{formatTime(localElapsed)}</span>
        <span>{currentSong ? formatTime(currentSong.duration) : "0:00"}</span>
      </div>
    </div>
  );
};

export const RadioPlayer = () => {
  const {
    currentSongFile,
    queueInfo,
    volume,
    isMuted,
    isAudioLoading,
    isAudioContextReady,
    isPlaying,
    startPlayback,
    getCurrentSong,
    calculateElapsedTime,
    handleVolumeChange,
    toggleMute,
    } = useRadio();

  const [isCompactMode, setIsCompactMode] = useState(false);



  const toggleCompactMode = () => {
    setIsCompactMode(!isCompactMode);
  };

  const currentSong = getCurrentSong();
  
  useEffect(() => {
    if (currentSongFile) {
      if (!currentSong) return;
      const elapsed = calculateElapsedTime(queueInfo?.StartTime || new Date().toISOString(), currentSong.duration);
      startPlayback(currentSongFile, elapsed);
    }
  }, [currentSongFile, startPlayback]);

  return (
    <>
      <VissonanceVisualizer />
      <VisualizerToggle />
      <VissonancePresetSelector />
      <ReactionBar />
      <div className={`
        ${isCompactMode 
          ? 'fixed bottom-4 left-4 right-4 sm:bottom-6 sm:left-20 sm:right-20 z-50 bg-black border border-gray-800 shadow-2xl p-2 sm:p-3 rounded-sm' 
          : 'w-full max-w-4xl mx-auto p-3 sm:p-4 bg-black border border-gray-800 shadow-2xl z-50'
        }
      `}>
        {/* Compact Mode Toggle */}
        <div className={`flex ${isCompactMode ? 'justify-between items-center' : 'justify-end'} mb-2`}>
          {isCompactMode && currentSong && (
            <div className="flex-1 min-w-0 mr-2 overflow-hidden">
              <p className="text-xs sm:text-sm md:text-base text-white font-mono truncate">
                {currentSong.title}
              </p>
              {isAudioLoading && (
                <p className="text-xs text-white font-mono truncate">
                  [LOADING...]
                </p>
              )}
              {isPlaying && (
                <p className="text-xs text-green-500 font-mono truncate">[PLAYING]</p>
              )}
            </div>
          )}
          
          <button
            onClick={toggleCompactMode}
            className="p-1 text-gray-600 hover:text-white transition-colors border border-gray-700 hover:border-white rounded-sm flex-shrink-0"
            aria-label={isCompactMode ? "Expand player" : "Compact player"}
          >
            {isCompactMode ? (
              <ChevronUpIcon className="h-3 w-3 sm:h-4 sm:w-4" />
            ) : (
              <ChevronDownIcon className="h-3 w-3 sm:h-4 sm:w-4" />
            )}
          </button>
        </div>

        {!isCompactMode && (
          <>
            {/* Header */}
            <div className="text-center mb-4 sm:mb-6">
              {!isAudioContextReady && (
                <p className="text-xs text-yellow-500 mb-3 font-mono">
                  [INITIALIZING AUDIO SYSTEM...]
                </p>
              )}
            </div>

            {/* Current Song Info */}
            <div className="text-center mb-4 sm:mb-6">
              {currentSong ? (
                <>
                  <p className="text-sm sm:text-lg text-white font-mono truncate px-2">
                    {currentSong.title}
                  </p>
                  {isAudioLoading && (
                    <p className="text-xs text-white font-mono mt-1">
                      [LOADING AUDIO...]
                    </p>
                  )}
                  {isPlaying && (
                    <p className="text-xs text-green-500 font-mono mt-1">[PLAYING]</p>
                  )}
                </>
              ) : (
                <div className="bg-gray-900 border border-gray-700 p-3 sm:p-4 mx-2 sm:mx-0">
                  <p className="text-sm sm:text-lg text-gray-500 font-mono">
                    [NO TRACK ACTIVE]
                  </p>
                </div>
              )}
            </div>
          </>
        )}

        {/* Progress Bar */}
        <div className={isCompactMode ? 'mb-1' : 'mb-2'}>
          <ProgressBar
            currentSong={currentSong}
            queueInfo={queueInfo}
          />
        </div>

        {/* Volume Controls */}
        <div className={`flex items-center gap-2 sm:gap-4 ${isCompactMode ? 'mb-0' : 'mb-2'} ${isCompactMode ? 'px-0' : 'px-2 sm:px-0'}`}>
          <button
            onClick={toggleMute}
            className="p-1 text-gray-600 hover:text-white transition-colors border border-gray-700 hover:border-white rounded-sm flex-shrink-0"
            aria-label={isMuted ? "Unmute" : "Mute"}
            disabled={!isAudioContextReady}
          >
            {isMuted || volume === 0 ? (
              <SpeakerXMarkIcon className={`${isCompactMode ? 'h-3 w-3' : 'h-3 w-3 sm:h-4 sm:w-4'}`} />
            ) : (
              <SpeakerWaveIcon className={`${isCompactMode ? 'h-3 w-3' : 'h-3 w-3 sm:h-4 sm:w-4'}`} />
            )}
          </button>

          <div className="flex flex-1 items-center min-w-0">
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

          {/* Volume percentage display on larger screens */}
          {!isCompactMode && (
            <span className="hidden sm:block text-xs text-gray-500 font-mono flex-shrink-0 w-8 text-right">
              {Math.round((isMuted ? 0 : volume) * 100)}%
            </span>
          )}
        </div>

        {/* Queue Info - Only shown in full mode */}
        {!isCompactMode && queueInfo?.Queue && queueInfo.Queue.length > 0 && (
          <div className="mt-4 sm:mt-6">
            <h3 className="text-xs sm:text-sm text-gray-500 font-mono mb-2">[QUEUE]</h3>
            <div className="space-y-1 max-h-32 sm:max-h-40 overflow-y-auto overflow-x-hidden">
              {/* Show current song */}
              {currentSong && (
                <div className="text-xs font-mono text-white truncate">
                  ▶ {currentSong.title}
                </div>
              )}
              {/* Show next 2 upcoming songs */}
              {queueInfo.Queue.slice(queueInfo.CurrentSongIndex + 1, queueInfo.CurrentSongIndex + 3).map((song, index) => (
                <div
                  key={song.youtube_id}
                  className="text-xs font-mono text-gray-500 truncate"
                >
                  {index + 2}. {song.title}
                </div>
              ))}
              {/* If we're near the end of the queue, show songs from the beginning */}
              {queueInfo.CurrentSongIndex + 3 > queueInfo.Queue.length && 
                queueInfo.Queue.slice(0, Math.min(2, queueInfo.Queue.length - 1)).map((song, index) => (
                  <div
                    key={`loop-${song.youtube_id}`}
                    className="text-xs font-mono text-gray-500 truncate"
                  >
                    {queueInfo.CurrentSongIndex + 2 + index}. {song.title}
                  </div>
                ))
              }
            </div>
            {queueInfo.Queue.length > 3 && (
              <div className="text-xs font-mono text-gray-600 mt-2 truncate">
                ... and {queueInfo.Queue.length - 3} more tracks
              </div>
            )}
          </div>
        )}
      </div>
    </>
  );
};
