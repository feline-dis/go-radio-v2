import { useState } from "react";
import { useReactions } from "../contexts/ReactionContext";
import { AnimatedEmotes } from "./AnimatedEmotes";

const EMOTES = [
  { id: "heart", emoji: "❤️", label: "Love" },
  { id: "fire", emoji: "🔥", label: "Fire" },
  { id: "rocket", emoji: "🚀", label: "Rocket" },
  { id: "clap", emoji: "👏", label: "Clap" },
  { id: "dance", emoji: "💃", label: "Dance" },
  { id: "party", emoji: "🎉", label: "Party" },
  { id: "star", emoji: "⭐", label: "Star" },
  { id: "thumbsup", emoji: "👍", label: "Thumbs Up" },
];

export const ReactionBar = () => {
  const { sendReaction, isWebSocketConnected, } = useReactions();
  const [selectedEmote, setSelectedEmote] = useState<string | null>(null);
  const [isOpen, setIsOpen] = useState(false);

  const handleEmoteClick = (emoteId: string) => {
    setSelectedEmote(emoteId);

    sendReaction(emoteId);

    // Reset selection after a short delay
    setTimeout(() => setSelectedEmote(null), 200);
  };

  const toggleReactionBar = () => {
    setIsOpen(!isOpen);
  };

  return (
    <>
    <AnimatedEmotes />
    <div className="fixed bottom-4 right-4 sm:bottom-6 sm:right-6 z-50">
      {/* Reaction Panel - extends upward when open */}
      {isOpen && (
        <div className="absolute bottom-12 sm:bottom-16 right-0 bg-black border border-gray-800 p-2 sm:p-4 rounded-sm shadow-2xl mb-2 w-44 sm:min-w-[200px]">
          <h3 className="text-xs sm:text-sm text-gray-500 font-mono mb-2 sm:mb-3 text-center">[REACTIONS]</h3>
          <div className="grid grid-cols-2 sm:flex sm:flex-col gap-1 sm:gap-2">
            {EMOTES.map((emote) => (
              <button
                key={emote.id}
                onClick={() => handleEmoteClick(emote.id)}
                className={`
                  p-2 sm:p-3 rounded-sm border transition-all duration-200 flex items-center justify-center sm:justify-start gap-1 sm:gap-3
                  ${
                    selectedEmote === emote.id
                      ? "border-white bg-white text-black"
                      : "border-gray-700 hover:border-gray-500 text-white hover:bg-gray-800"
                  }
                `}
                title={emote.label}
                disabled={!isWebSocketConnected}
              >
                <span className="text-sm sm:text-lg">{emote.emoji}</span>
                <span className="text-xs sm:text-sm font-mono hidden sm:inline">{emote.label}</span>
              </button>
            ))}
          </div>
          {!isWebSocketConnected && (
            <p className="text-xs text-red-500 font-mono mt-2 sm:mt-3 text-center">
              [DISCONNECTED]
            </p>
          )}
        </div>
      )}

      {/* Floating Action Button */}
      <button
        onClick={toggleReactionBar}
        className={`
          w-10 h-10 sm:w-12 sm:h-12 border-2 transition-all duration-200 flex items-center justify-center rounded-sm sm:rounded-none
          ${
            isOpen
              ? "bg-white border-white text-black rotate-45"
              : "bg-black border-gray-700 text-white hover:border-white hover:bg-gray-900"
          }
          ${!isWebSocketConnected ? "opacity-50 cursor-not-allowed" : ""}
        `}
        disabled={!isWebSocketConnected}
        title={isOpen ? "Close reactions" : "Send reaction"}
      >
        {isOpen ? (
          <span className="text-lg sm:text-2xl font-bold">+</span>
        ) : (
          <span className="text-lg sm:text-2xl">😊</span>
        )}
      </button>
    </div>
    
    </>
  );
}; 