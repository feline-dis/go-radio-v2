import { useNavigate } from "react-router-dom";
import { useRadio } from "../contexts/RadioContext";

export const IntroPage = () => {
  const navigate = useNavigate();
  const {
    isQueueLoading,
    isWebSocketConnected,
    queueInfo,
    setIsUserInteracted,
    initAudioContext,
  } = useRadio();

  const handleEnterRadio = () => {
    setIsUserInteracted(true);
    initAudioContext();
    navigate("/player");
  };

  const currentSong = queueInfo.Queue?.[queueInfo.CurrentSongIndex];

  return (
    <div className="w-full max-w-md mx-auto p-4 sm:p-8 bg-black border border-gray-800 shadow-2xl">
      <div className="text-center">
        <h2 className="text-lg sm:text-xl font-mono font-bold text-white mb-4 sm:mb-6 tracking-wider">
          GO_RADIO
        </h2>

        {/* Current Song Display */}
        {currentSong && (
          <div className="mb-4 sm:mb-6 p-3 border border-gray-700 bg-gray-900">
            <p className="text-xs text-gray-500 font-mono mb-1">[NOW PLAYING]</p>
            <p className="text-sm text-white font-mono truncate">
              {currentSong.title}
            </p>
          </div>
        )}

        {isQueueLoading && (
          <div className="mb-4 sm:mb-6">
            <p className="text-gray-500 mb-3 font-mono text-xs sm:text-sm">
              [LOADING QUEUE DATA...]
            </p>
            <div className="w-5 h-5 sm:w-6 sm:h-6 border-2 border-white border-t-transparent rounded-none animate-spin mx-auto"></div>
          </div>
        )}

        {!isWebSocketConnected && !isQueueLoading && (
          <div className="mb-4 sm:mb-6">
            <p className="text-gray-500 mb-3 font-mono text-xs sm:text-sm">
              [CONNECTING TO SERVER...]
            </p>
            <div className="w-5 h-5 sm:w-6 sm:h-6 border-2 border-white border-t-transparent rounded-none animate-spin mx-auto"></div>
          </div>
        )}

        {!isQueueLoading && isWebSocketConnected && queueInfo.Queue?.length > 0 && (
          <div className="mb-6 sm:mb-8">
            <p className="text-white mb-4 sm:mb-6 font-mono text-sm">[SYSTEM READY]</p>
            <button
              onClick={handleEnterRadio}
              className="px-6 sm:px-8 py-2 sm:py-3 bg-black border border-white hover:bg-white hover:text-black text-white font-mono transition-colors text-sm sm:text-base"
            >
              [ENTER RADIO]
            </button>
            <p className="text-xs text-gray-500 mt-3 font-mono">
              [INITIALIZES AUDIO SYSTEM]
            </p>
          </div>
        )}

        {/* Status indicators */}
        <div className="space-y-3 text-xs font-mono">
          <div
            className={`flex items-center justify-center gap-3 ${
              !isQueueLoading ? "text-white" : "text-gray-500"
            }`}
          >
            <div
              className={`w-2 h-2 border ${
                !isQueueLoading ? "bg-white border-white" : "border-gray-500"
              }`}
            ></div>
            QUEUE {!isQueueLoading ? "[LOADED]" : "[LOADING...]"}
          </div>
          <div
            className={`flex items-center justify-center gap-3 ${
              isWebSocketConnected ? "text-white" : "text-gray-500"
            }`}
          >
            <div
              className={`w-2 h-2 border ${
                isWebSocketConnected
                  ? "bg-white border-white"
                  : "border-gray-500"
              }`}
            ></div>
            WEBSOCKET {isWebSocketConnected ? "[CONNECTED]" : "[CONNECTING...]"}
          </div>
          <div
            className={`flex items-center justify-center gap-3 ${
              queueInfo.Queue?.length > 0 ? "text-white" : "text-gray-500"
            }`}
          >
            <div
              className={`w-2 h-2 border ${
                queueInfo.Queue?.length > 0
                  ? "bg-white border-white"
                  : "border-gray-500"
              }`}
            ></div>
            AUDIO {queueInfo.Queue?.length > 0 ? "[READY]" : "[LOADING...]"}
          </div>
        </div>

        {/* Queue Info */}
        {queueInfo.Queue?.length > 0 && (
          <div className="mt-6 pt-4 border-t border-gray-700">
            <p className="text-xs text-gray-500 font-mono mb-2">[QUEUE INFO]</p>
            <div className="text-xs text-gray-400 font-mono space-y-1">
              <div>TRACKS: {queueInfo.Queue.length}</div>
              {queueInfo.Playlist && (
                <div>PLAYLIST: {queueInfo.Playlist.name}</div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};