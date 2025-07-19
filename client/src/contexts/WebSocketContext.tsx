import React, { createContext, useContext, useEffect, useRef, useState, useCallback } from 'react';
import type { Playlist, Song } from './RadioContext';

// Event types that can be sent/received via WebSocket
export interface WebSocketEvents {
  song_change: {
    current_song: Song;
    next_song: Song;
    playlist: Playlist;
    queue: Song[];
    remaining: number;
    start_time: string;
    current_song_index: number;
  };
  queue_update: {
    queue: Song[];
    playlist: Playlist;
    remaining: number;
    start_time: string;
    current_song_index: number;
  };
  playback_update: {
    current_time: number;
    duration: number;
    is_playing: boolean;
  };
  user_reaction: {
    user_id: string;
    emote: string;
    timestamp: number;
  };
  ping: {};
  pong: {};
  get_playback_state: {};
}

// WebSocket message structure
interface WebSocketMessage<T extends keyof WebSocketEvents = keyof WebSocketEvents> {
  type: T;
  payload: WebSocketEvents[T];
  timestamp?: number;
}

// Event handler function type
type EventHandler<T extends keyof WebSocketEvents> = (data: WebSocketEvents[T]) => void;

// Event bus interface
interface EventBus {
  subscribe<T extends keyof WebSocketEvents>(event: T, handler: EventHandler<T>): () => void;
  publish<T extends keyof WebSocketEvents>(event: T, data: WebSocketEvents[T]): void;
  unsubscribe<T extends keyof WebSocketEvents>(event: T, handler: EventHandler<T>): void;
}

// WebSocket connection states
export enum ConnectionState {
  CONNECTING = 'connecting',
  CONNECTED = 'connected',
  DISCONNECTED = 'disconnected',
  RECONNECTING = 'reconnecting',
  ERROR = 'error'
}

// WebSocket context interface
interface WebSocketContextType {
  connectionState: ConnectionState;
  eventBus: EventBus;
  sendMessage<T extends keyof WebSocketEvents>(type: T, data: WebSocketEvents[T]): void;
  connect: () => void;
  disconnect: () => void;
  isConnected: boolean;
}

// Create the context
const WebSocketContext = createContext<WebSocketContextType | null>(null);

// Event bus implementation
class ClientEventBus implements EventBus {
  private handlers: Map<keyof WebSocketEvents, Set<EventHandler<any>>> = new Map();

  subscribe<T extends keyof WebSocketEvents>(event: T, handler: EventHandler<T>): () => void {
    if (!this.handlers.has(event)) {
      this.handlers.set(event, new Set());
    }
    this.handlers.get(event)!.add(handler);

    // Return unsubscribe function
    return () => this.unsubscribe(event, handler);
  }

  publish<T extends keyof WebSocketEvents>(event: T, data: WebSocketEvents[T]): void {
    const eventHandlers = this.handlers.get(event);
    if (eventHandlers) {
      eventHandlers.forEach(handler => {
        try {
          handler(data);
        } catch (error) {
          console.error(`Error in event handler for ${event}:`, error);
        }
      });
    }
  }

  unsubscribe<T extends keyof WebSocketEvents>(event: T, handler: EventHandler<T>): void {
    const eventHandlers = this.handlers.get(event);
    if (eventHandlers) {
      eventHandlers.delete(handler);
    }
  }

  // Clear all handlers for cleanup
  clear(): void {
    this.handlers.clear();
  }
}

// WebSocket provider component
interface WebSocketProviderProps {
  children: React.ReactNode;
  url?: string;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({
  children,
  url = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`,
  reconnectInterval = 3000,
  maxReconnectAttempts = 10
}) => {
  const [connectionState, setConnectionState] = useState<ConnectionState>(ConnectionState.DISCONNECTED);
  const websocketRef = useRef<WebSocket | null>(null);
  const eventBusRef = useRef<ClientEventBus>(new ClientEventBus());
  const reconnectAttemptsRef = useRef(0);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const pingIntervalRef = useRef<number | null>(null);

  // Send ping every 30 seconds to keep connection alive
  const startPingInterval = useCallback(() => {
    if (pingIntervalRef.current) {
      clearInterval(pingIntervalRef.current);
    }
    pingIntervalRef.current = setInterval(() => {
      if (websocketRef.current?.readyState === WebSocket.OPEN) {
        sendMessage('ping', {});
      }
    }, 30000);
  }, []);

  const stopPingInterval = useCallback(() => {
    if (pingIntervalRef.current) {
      clearInterval(pingIntervalRef.current);
      pingIntervalRef.current = null;
    }
  }, []);

  const connect = useCallback(() => {
    if (websocketRef.current?.readyState === WebSocket.OPEN) {
      return; // Already connected
    }

    try {
      setConnectionState(ConnectionState.CONNECTING);
      const ws = new WebSocket(url);

      ws.onopen = () => {
        setConnectionState(ConnectionState.CONNECTED);
        reconnectAttemptsRef.current = 0;
        startPingInterval();

        // Request current playback state when connected
        sendMessage('get_playback_state', {});
      };

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);

          // Handle pong messages
          if (message.type === 'pong') {
            return;
          }

          // Publish the event through the event bus
          eventBusRef.current.publish(message.type, message.payload);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      ws.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        setConnectionState(ConnectionState.DISCONNECTED);
        stopPingInterval();

        // Attempt to reconnect if not manually closed
        if (event.code !== 1000 && reconnectAttemptsRef.current < maxReconnectAttempts) {
          setConnectionState(ConnectionState.RECONNECTING);
          reconnectTimeoutRef.current = setTimeout(() => {
            reconnectAttemptsRef.current++;
            connect();
          }, reconnectInterval);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setConnectionState(ConnectionState.ERROR);
      };

      websocketRef.current = ws;
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      setConnectionState(ConnectionState.ERROR);
    }
  }, [url, reconnectInterval, maxReconnectAttempts, startPingInterval, stopPingInterval]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    stopPingInterval();

    if (websocketRef.current) {
      websocketRef.current.close(1000, 'Manual disconnect');
      websocketRef.current = null;
    }

    setConnectionState(ConnectionState.DISCONNECTED);
    reconnectAttemptsRef.current = 0;
  }, [stopPingInterval]);

  const sendMessage = useCallback(<T extends keyof WebSocketEvents>(
    type: T,
    data: WebSocketEvents[T]
  ) => {
    if (websocketRef.current?.readyState === WebSocket.OPEN) {
      const message: WebSocketMessage<T> = {
        type,
        payload: data,
        timestamp: Date.now()
      };
      websocketRef.current.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket not connected, cannot send message:', type);
    }
  }, []);

  // Auto-connect on mount
  useEffect(() => {
    connect();

    return () => {
      disconnect();
      eventBusRef.current.clear();
    };
  }, [connect, disconnect]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      stopPingInterval();
    };
  }, [stopPingInterval]);

  const contextValue: WebSocketContextType = {
    connectionState,
    eventBus: eventBusRef.current,
    sendMessage,
    connect,
    disconnect,
    isConnected: connectionState === ConnectionState.CONNECTED
  };

  return (
    <WebSocketContext.Provider value={contextValue}>
      {children}
    </WebSocketContext.Provider>
  );
};

// Hook to use the WebSocket context
export const useWebSocket = (): WebSocketContextType => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};

// Utility hook for subscribing to specific events
export const useWebSocketEvent = <T extends keyof WebSocketEvents>(
  event: T,
  handler: EventHandler<T>,
  deps: React.DependencyList = []
) => {
  const { eventBus } = useWebSocket();

  useEffect(() => {
    const unsubscribe = eventBus.subscribe(event, handler);
    return unsubscribe;
  }, [eventBus, event, ...deps]);
};

// Utility hook for publishing events
export const useWebSocketPublish = () => {
  const { eventBus, sendMessage } = useWebSocket();

  return {
    publish: eventBus.publish.bind(eventBus),
    sendMessage
  };
}; 
