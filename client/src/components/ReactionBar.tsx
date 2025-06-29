import { useState } from "react";
import { useReactions } from "../contexts/ReactionContext";

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
  const { sendReaction, isWebSocketConnected } = useReactions();
  const [selectedEmote, setSelectedEmote] = useState<string | null>(null);
  const [isOpen, setIsOpen] = useState(false);

  const handleEmoteClick = (emoteId: string) => {
    setSelectedEmote(emoteId);

    // Generate a simple user ID for demo purposes
    // In a real app, this would come from user authentication
    const userId = `user_${Math.random().toString(36).substr(2, 9)}`;

    sendReaction(userId, emoteId);

    // Reset selection after a short delay
    setTimeout(() => setSelectedEmote(null), 200);
    
    // Close the reaction bar after sending a reaction
    setIsOpen(false);
  };

  const toggleReactionBar = () => {
    setIsOpen(!isOpen);
  };

  return (
    <div className="fixed bottom-6 right-6 z-50">
      {/* Reaction Panel - extends upward when open */}
      {isOpen && (
        <div className="absolute bottom-16 right-0 bg-black border border-gray-800 p-4 rounded-sm shadow-2xl mb-2 min-w-[200px]">
          <h3 className="text-sm text-gray-500 font-mono mb-3 text-center">[REACTIONS]</h3>
          <div className="flex flex-col gap-2">
            {EMOTES.map((emote) => (
              <button
                key={emote.id}
                onClick={() => handleEmoteClick(emote.id)}
                className={`
                  p-3 rounded-sm border transition-all duration-200 flex items-center gap-3
                  ${
                    selectedEmote === emote.id
                      ? "border-white bg-white text-black"
                      : "border-gray-700 hover:border-gray-500 text-white hover:bg-gray-800"
                  }
                `}
                title={emote.label}
                disabled={!isWebSocketConnected}
              >
                <span className="text-lg">{emote.emoji}</span>
                <span className="text-sm font-mono">{emote.label}</span>
              </button>
            ))}
          </div>
          {!isWebSocketConnected && (
            <p className="text-xs text-red-500 font-mono mt-3 text-center">
              [DISCONNECTED]
            </p>
          )}
        </div>
      )}

      {/* Floating Action Button */}
      <button
        onClick={toggleReactionBar}
        className={`
          w-14 h-14 rounded-full border-2 transition-all duration-200 flex items-center justify-center
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
          <span className="text-2xl font-bold">+</span>
        ) : (
          <span className="text-2xl">-</span>
        )}
      </button>
    </div>
  );
};
