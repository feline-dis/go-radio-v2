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

  const handleEmoteClick = (emoteId: string) => {
    setSelectedEmote(emoteId);

    // Generate a simple user ID for demo purposes
    // In a real app, this would come from user authentication
    const userId = `user_${Math.random().toString(36).substr(2, 9)}`;

    sendReaction(userId, emoteId);

    // Reset selection after a short delay
    setTimeout(() => setSelectedEmote(null), 200);
  };

  return (
    <div className="bg-black border border-gray-800 p-4 rounded-sm">
      <h3 className="text-sm text-gray-500 font-mono mb-3">[REACTIONS]</h3>
      <div className="flex gap-2 flex-wrap">
        {EMOTES.map((emote) => (
          <button
            key={emote.id}
            onClick={() => handleEmoteClick(emote.id)}
            className={`
              p-2 rounded-sm border transition-all duration-200
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
          </button>
        ))}
      </div>
      {!isWebSocketConnected && (
        <p className="text-xs text-red-500 font-mono mt-2">
          [DISCONNECTED - REACTIONS UNAVAILABLE]
        </p>
      )}
    </div>
  );
};
