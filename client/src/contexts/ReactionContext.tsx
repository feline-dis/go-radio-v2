import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useRef,
  useCallback,
} from "react";
import { useWebSocket, useWebSocketEvent, useWebSocketPublish } from "./WebSocketContext";

interface EmoteDisplay {
  id: string;
  emoji: string;
  x: number;
  y: number;
  timestamp: number;
}

interface ReactionEventPayload {
  emote: string;
  timestamp: number;
}

interface ReactionContextType {
  // State
  isWebSocketConnected: boolean;
  reactions: EmoteDisplay[];

  // Functions
  sendReaction: (emote: string) => void;
}

const ReactionContext = createContext<ReactionContextType | undefined>(
  undefined
);

export const useReactions = () => {
  const context = useContext(ReactionContext);
  if (context === undefined) {
    throw new Error("useReactions must be used within a ReactionProvider");
  }
  return context;
};

interface ReactionProviderProps {
  children: React.ReactNode;
}

const EMOTE_MAP: Record<string, string> = {
  heart: "❤️",
  fire: "🔥",
  rocket: "🚀",
  clap: "👏",
  dance: "💃",
  party: "🎉",
  star: "⭐",
  thumbsup: "👍",
};

export const ReactionProvider: React.FC<ReactionProviderProps> = ({
  children,
}) => {
  // Use centralized WebSocket connection
  const { isConnected: isWebSocketConnected } = useWebSocket();
  const { sendMessage } = useWebSocketPublish();
  
  const [reactions, setReactions] = useState<EmoteDisplay[]>([]);
  const reactionTimeoutRefs = useRef<Map<string, number>>(new Map());

  // Clean up reactions after animation
  const cleanupReaction = useCallback((reactionId: string) => {
    setReactions((prev) => prev.filter((r) => r.id !== reactionId));
    reactionTimeoutRefs.current.delete(reactionId);
  }, []);

  // Handle incoming reaction events
  const handleReactionEvent = useCallback(
    (payload: ReactionEventPayload) => {
      console.log("🎭 Received reaction:", payload);
      const {  emote, timestamp } = payload;

      // Generate random position for the emote
      const x = Math.random() * 80 + 10; // 10% to 90% of screen width
      const y = Math.random() * 60 + 20; // 20% to 80% of screen height

      const reactionId = `${timestamp}`;
      const emoji = EMOTE_MAP[emote] || "👍";

      const newReaction: EmoteDisplay = {
        id: reactionId,
        emoji,
        x,
        y,
        timestamp,
      };

      setReactions((prev) => [...prev, newReaction]);

      // Clean up reaction after 3 seconds
      const timeout = setTimeout(() => {
        cleanupReaction(reactionId);
      }, 3000);

      reactionTimeoutRefs.current.set(reactionId, timeout);
    },
    [cleanupReaction]
  );

  // Send reaction function using centralized WebSocket
  const sendReaction = useCallback((emote: string) => {
    if (!isWebSocketConnected) {
      console.warn("WebSocket not connected, cannot send reaction");
      return;
    }

    try {
      const reactionData = {
        user_id: `user_${Date.now()}`, // TODO: Get actual user ID from auth context
        emote: emote,
        timestamp: Date.now(),
      };
      console.log("🎭 Sending reaction:", reactionData);
      sendMessage('user_reaction', reactionData);
      console.log("🎭 Reaction sent successfully");
    } catch (error) {
      console.error("Failed to send reaction:", error);
    }
  }, [isWebSocketConnected, sendMessage]);

  // Subscribe to user_reaction events from centralized WebSocket
  useWebSocketEvent('user_reaction', (data) => {
    handleReactionEvent(data);
  }, [handleReactionEvent]);

  // Cleanup reactions on unmount
  useEffect(() => {
    return () => {
      // Clean up all reaction timeouts
      reactionTimeoutRefs.current.forEach((timeout) => clearTimeout(timeout));
      reactionTimeoutRefs.current.clear();
    };
  }, []);

  const value: ReactionContextType = {
    isWebSocketConnected,
    reactions,
    sendReaction,
  };

  return (
    <ReactionContext.Provider value={value}>
      {children}
    </ReactionContext.Provider>
  );
};
