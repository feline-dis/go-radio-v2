import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useRef,
  useCallback,
} from "react";

interface EmoteDisplay {
  id: string;
  emoji: string;
  x: number;
  y: number;
  timestamp: number;
}

interface ReactionEventPayload {
  user_id: string;
  emote: string;
  timestamp: number;
}

interface ReactionContextType {
  // State
  isWebSocketConnected: boolean;
  reactions: EmoteDisplay[];

  // Functions
  sendReaction: (userId: string, emote: string) => void;
  connectWebSocket: (wsUrl: string) => void;
  disconnectWebSocket: () => void;
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
  wsUrl: string;
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
  wsUrl,
}) => {
  const [isWebSocketConnected, setIsWebSocketConnected] = useState(false);
  const [reactions, setReactions] = useState<EmoteDisplay[]>([]);

  const wsRef = useRef<WebSocket | null>(null);
  const reactionTimeoutRefs = useRef<Map<string, number>>(new Map());

  // Clean up reactions after animation
  const cleanupReaction = useCallback((reactionId: string) => {
    setReactions((prev) => prev.filter((r) => r.id !== reactionId));
    reactionTimeoutRefs.current.delete(reactionId);
  }, []);

  // Handle incoming reaction events
  const handleReactionEvent = useCallback(
    (payload: ReactionEventPayload) => {
      const { user_id, emote, timestamp } = payload;

      // Generate random position for the emote
      const x = Math.random() * 80 + 10; // 10% to 90% of screen width
      const y = Math.random() * 60 + 20; // 20% to 80% of screen height

      const reactionId = `${user_id}_${timestamp}`;
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

  // Send reaction function
  const sendReaction = useCallback((userId: string, emote: string) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.warn("WebSocket not connected, cannot send reaction");
      return;
    }

    const reactionMessage = {
      type: "reaction",
      user_id: userId,
      emote: emote,
    };

    try {
      wsRef.current.send(JSON.stringify(reactionMessage));
      console.log("🎭 Sent reaction:", emote);
    } catch (error) {
      console.error("Failed to send reaction:", error);
    }
  }, []);

  // Connect to WebSocket
  const connectWebSocket = useCallback(
    (url: string) => {
      console.log("Connecting to reaction WebSocket at:", url);
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log("✅ Connected to reaction WebSocket successfully");
        setIsWebSocketConnected(true);
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);

          switch (data.type) {
            case "user_reaction":
              handleReactionEvent(data.payload);
              break;
            case "pong":
              console.log("🏓 Received pong from reaction server");
              break;
            default:
              // Ignore other message types (they're handled by RadioContext)
              break;
          }
        } catch (error) {
          console.error(
            "❌ Failed to parse reaction WebSocket message:",
            error
          );
        }
      };

      ws.onerror = (error) => {
        console.error("❌ Reaction WebSocket error:", error);
        setIsWebSocketConnected(false);
      };

      ws.onclose = (event) => {
        console.log(
          "🔌 Disconnected from reaction WebSocket. Code:",
          event.code
        );
        setIsWebSocketConnected(false);
        // Attempt to reconnect after 5 seconds
        setTimeout(() => connectWebSocket(url), 5000);
      };
    },
    [handleReactionEvent]
  );

  // Disconnect WebSocket
  const disconnectWebSocket = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsWebSocketConnected(false);
  }, []);

  // Connect on mount
  useEffect(() => {
    connectWebSocket(wsUrl);

    return () => {
      disconnectWebSocket();
      // Clean up all reaction timeouts
      reactionTimeoutRefs.current.forEach((timeout) => clearTimeout(timeout));
      reactionTimeoutRefs.current.clear();
    };
  }, [wsUrl, connectWebSocket, disconnectWebSocket]);

  const value: ReactionContextType = {
    isWebSocketConnected,
    reactions,
    sendReaction,
    connectWebSocket,
    disconnectWebSocket,
  };

  return (
    <ReactionContext.Provider value={value}>
      {children}
    </ReactionContext.Provider>
  );
};
