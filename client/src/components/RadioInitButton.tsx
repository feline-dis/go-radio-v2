import { useState } from "react";
import { useRadio } from "../contexts/NewRadioContext";
import { RadioPlayer } from "./RadioPlayer";

export const RadioInitButton = () => {
  const {
    isReady,
    isQueueLoading,
    isWebSocketConnected,
    setIsUserInteracted,
    initAudioContext,
  } = useRadio();
  const [showPlayer, setShowPlayer] = useState(false);

  const handleEnterRadio = () => {
    setIsUserInteracted(true);
    initAudioContext();
    setShowPlayer(true);
  };

  // Show the radio player if we're ready and user has clicked enter
  if (showPlayer && isReady) {
    return <RadioPlayer />;
  }

  // Show initialization button
  return (
    <div className="w-full max-w-md mx-auto p-4 sm:p-8 bg-black border border-gray-800 shadow-2xl">
      <div className="text-center">
        <h2 className="text-lg sm:text-xl font-mono font-bold text-white mb-4 sm:mb-6 tracking-wider">
          GO_RADIO
        </h2>

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

        {!isQueueLoading && isWebSocketConnected && (
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
        </div>
      </div>
    </div>
  );
};
