import { BrowserRouter as Router, Routes, Route, Link } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactionProvider } from "./contexts/ReactionContext";
import { AuthProvider, useAuth } from "./contexts/AuthContext";
import { VisualizerProvider } from "./contexts/VisualizerContext";
import { WebSocketProvider } from "./contexts/WebSocketContext";
import { IntroPage } from "./pages/IntroPage";
import { PlayerPage } from "./pages/PlayerPage";
import { CreatePlaylist } from "./pages/CreatePlaylist";
import { LoginPage } from "./pages/LoginPage";
import { AdminPage } from "./pages/AdminPage";
import { LogoutButton } from "./components/LogoutButton";
import { ReactionBar } from "./components/ReactionBar";
import { AnimatedEmotes } from "./components/AnimatedEmotes";
import { Toaster } from "react-hot-toast";
import { RadioProvider } from "./contexts/RadioContext";

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 0, // Don't cache real-time data
      retry: 1,
      refetchOnWindowFocus: false, // Disable refetch on window focus globally
      refetchOnMount: true, // Always refetch on mount
      refetchOnReconnect: true, // Refetch on reconnect
    },
  },
});

function AppContent() {
  const { isAuthenticated, user } = useAuth();

  return (
    <Router>
      <div className="min-h-screen bg-black flex flex-col relative">
        {/* Navigation */}
        <nav className="bg-black border-b border-gray-800 relative z-10">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between h-16">
              <div className="flex">
                <div className="flex-shrink-0 flex items-center">
                  <Link
                    to="/"
                    className="text-xl font-mono font-bold text-white tracking-wider"
                  >
                    GO_RADIO
                  </Link>
                </div>
                <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
                  <Link
                    to="/player"
                    className="border-transparent text-gray-500 hover:text-white hover:border-white inline-flex items-center px-1 pt-1 border-b-2 text-sm font-mono transition-colors"
                  >
                    [PLAYER]
                  </Link>
                  {isAuthenticated && (
                    <Link
                      to="/playlists/create"
                      className="border-transparent text-gray-500 hover:text-white hover:border-white inline-flex items-center px-1 pt-1 border-b-2 text-sm font-mono transition-colors"
                    >
                      [CREATE_PLAYLIST]
                    </Link>
                  )}


                  {isAuthenticated && (
                    <Link
                      to="/admin"
                      className="border-transparent text-gray-500 hover:text-white hover:border-white inline-flex items-center px-1 pt-1 border-b-2 text-sm font-mono transition-colors"
                    >
                      [ADMIN]
                    </Link>
                  )}
                </div>
              </div>
              <div className="flex items-center space-x-4">
                {isAuthenticated ? (
                  <div className="flex items-center space-x-4">
                    <span className="text-gray-400 font-mono text-sm">
                      {user?.username}
                    </span>
                    <LogoutButton variant="icon" />
                  </div>
                ) : (
                  <Link
                    to="/login"
                    className="text-gray-500 hover:text-white font-mono text-sm transition-colors"
                  >
                    [LOGIN]
                  </Link>
                )}
              </div>
            </div>
          </div>
        </nav>

        {/* Main Content */}
        <main className="flex-1 flex items-center justify-center relative z-10">
          <Routes>
            <Route path="/" element={<IntroPage />} />
            <Route path="/player" element={<PlayerPage />} />
            <Route path="/playlists/create" element={<CreatePlaylist />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/admin" element={<AdminPage />} />
          </Routes>
        </main>

        {/* Animated Emotes and Reaction Bar - Only show on player page */}
        <AnimatedEmotes />
        <ReactionBar />
        
      </div>
    </Router>
  );
}

function App() {
  // In development, we'll connect to the local server
  const wsUrl = import.meta.env.DEV
    ? "ws://localhost:8080/ws"
    : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;

  return (
    <>
      <Toaster position="top-right" />
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <WebSocketProvider url={wsUrl}>
            <ReactionProvider>
              <RadioProvider>
                <VisualizerProvider>
                  <AppContent />
                </VisualizerProvider>
              </RadioProvider>
            </ReactionProvider>
          </WebSocketProvider>
        </AuthProvider>
      </QueryClientProvider>
    </>
  );
}

export default App;
