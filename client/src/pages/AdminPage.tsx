import React from "react";
import { useAuth } from "../contexts/AuthContext";
import { ProtectedRoute } from "../components/ProtectedRoute";
import { LogoutButton } from "../components/LogoutButton";
import { useRadio } from "../contexts/RadioContext";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "react-hot-toast";
import api from "../lib/axios";

interface Playlist {
  id: string;
  name: string;
  description: string;
  song_count: number;
  created_at: string;
  updated_at: string;
}

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

// Helper function to derive current song from queue and index
const getCurrentSong = (queueInfo: QueueInfo | null): Song | null => {
  if (!queueInfo || !queueInfo.Queue || queueInfo.Queue.length === 0) {
    return null;
  }
  
  const currentIndex = queueInfo.CurrentSongIndex;
  if (currentIndex < 0 || currentIndex >= queueInfo.Queue.length) {
    return null;
  }
  
  return queueInfo.Queue[currentIndex];
};

const AdminPageContent: React.FC = () => {
  const { user } = useAuth();
  const { queueInfo } = useRadio();
  const queryClient = useQueryClient();

  const handleSkip = async () => {
    try {
      await api.post("/admin/skip");
      toast.success("Skipped to next song");
    } catch (error: any) {
      toast.error(error.response?.data?.error || "Failed to skip song");
    }
  };

  const handlePrevious = async () => {
    try {
      await api.post("/admin/previous");
      toast.success("Previous song");
    } catch (error: any) {
      toast.error(error.response?.data?.error || "Failed to go to previous song");
    }
  };

  // Fetch all playlists
  const { data: playlists, isLoading: playlistsLoading, error: playlistsError } = useQuery<Playlist[]>({
    queryKey: ["playlists"],
    queryFn: async () => {
      const response = await api.get("/playlists");
      return response.data;
    },
    refetchOnWindowFocus: false,
  });

  // Set active playlist mutation
  const setActivePlaylistMutation = useMutation({
    mutationFn: async (playlistId: string) => {
      const response = await api.post("/admin/playlist/set-active", {
        playlist_id: playlistId,
      });
      return response.data;
    },
    onSuccess: (_, playlistId) => {
      const playlist = playlists?.find(p => p.id === playlistId);
      toast.success(`Switched to playlist: ${playlist?.name || 'Unknown'}`);
      // Invalidate and refetch related queries
      queryClient.invalidateQueries({ queryKey: ["queue"] });
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || "Failed to switch playlist");
    },
  });

  const handleSetActivePlaylist = (playlistId: string) => {
    setActivePlaylistMutation.mutate(playlistId);
  };

  // Format date for display
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  return (
    <div className="min-h-screen bg-black text-white p-4 md:p-8">
      <div className="max-w-6xl mx-auto">
        {/* Terminal Header */}
        <div className="bg-black border border-gray-700 mb-8">
          <div className="border-b border-gray-700 p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-4">
                <div className="flex space-x-2">
                  <div className="w-3 h-3 bg-red-500 border border-red-400"></div>
                  <div className="w-3 h-3 bg-yellow-500 border border-yellow-400"></div>
                  <div className="w-3 h-3 bg-green-500 border border-green-400"></div>
                </div>
                <span className="text-gray-400 font-mono text-sm">ADMIN_TERMINAL v2.1.0</span>
              </div>
              <LogoutButton />
            </div>
          </div>
          <div className="p-6">
            <div className="font-mono">
              <div className="text-green-400 mb-1">
                <span className="text-white">[ADMIN_CONTROL_INTERFACE]</span>
              </div>
              <div className="text-green-400 mb-1">
                <span className="text-gray-400">User:</span> <span className="text-white">{user?.username}</span>
              </div>
              <div className="text-green-400 mb-1">
                <span className="text-gray-400">Access Level:</span> <span className="text-yellow-400">ADMINISTRATOR</span>
              </div>
              <div className="text-green-400 mb-1">
                <span className="text-gray-400">Session:</span> <span className="text-green-400">ACTIVE</span>
              </div>
            </div>
          </div>
        </div>

        {/* Radio Controls */}
        <div className="bg-black border border-gray-700 mb-8">
          <div className="border-b border-gray-700 p-4">
            <h2 className="text-green-400 font-mono font-bold text-lg tracking-wider">
              ► RADIO_CONTROL_INTERFACE
            </h2>
          </div>
          <div className="p-6">
            <div className="grid grid-cols-2 gap-4">
              <button
                onClick={handlePrevious}
                className="bg-black border border-purple-500 hover:bg-purple-500 hover:text-black text-purple-400 hover:text-black font-mono py-4 px-4 transition-colors relative group"
              >
                <span>◄ PREV</span>
              </button>
              <button
                onClick={handleSkip}
                className="bg-black border border-blue-500 hover:bg-blue-500 hover:text-black text-blue-400 font-mono py-4 px-4 transition-colors relative group"
              >
                <span>SKIP ►</span>
              </button>
            </div>
          </div>
        </div>

        {/* Playlist Management */}
        <div className="bg-black border border-gray-700 mb-8">
          <div className="border-b border-gray-700 p-4">
            <h2 className="text-green-400 font-mono font-bold text-lg tracking-wider">
              ► PLAYLIST_DATABASE_INTERFACE
            </h2>
            <div className="text-gray-400 font-mono text-sm mt-1">
              {playlists ? `[${playlists.length} RECORDS FOUND]` : '[SCANNING...]'}
            </div>
          </div>
          
          <div className="p-6">
            {playlistsLoading && (
              <div className="text-center py-8">
                <div className="inline-flex items-center space-x-3">
                  <div className="w-4 h-4 border-2 border-green-400 border-t-transparent animate-spin"></div>
                  <p className="text-green-400 font-mono">[SCANNING PLAYLIST DATABASE...]</p>
                </div>
              </div>
            )}

            {playlistsError && (
              <div className="text-center py-8 bg-red-900/20 border border-red-700">
                <p className="text-red-400 font-mono">[ERROR] DATABASE CONNECTION FAILED</p>
                <p className="text-gray-400 font-mono text-sm mt-1">Unable to retrieve playlist data</p>
              </div>
            )}

            {playlists && playlists.length > 0 && (
              <div className="bg-black border border-gray-600">
                {/* Table Header */}
                <div className="bg-gray-800 border-b border-gray-600 p-3">
                  <div className="grid grid-cols-12 gap-4 font-mono text-xs text-green-400 font-bold">
                    <div className="col-span-3">PLAYLIST_NAME</div>
                    <div className="col-span-3">DESCRIPTION</div>
                    <div className="col-span-1">TRACKS</div>
                    <div className="col-span-2">CREATED</div>
                    <div className="col-span-1">STATUS</div>
                    <div className="col-span-2">ACTIONS</div>
                  </div>
                </div>
                
                {/* Table Body */}
                <div className="divide-y divide-gray-700">
                  {playlists.map((playlist, index) => {
                    const isActive = queueInfo?.Playlist?.id?.toString() === playlist.id;
                    
                    return (
                      <div 
                        key={playlist.id} 
                        className={`grid grid-cols-12 gap-4 p-3 font-mono text-sm hover:bg-gray-800/50 transition-colors ${
                          isActive ? 'bg-green-900/20 border-l-4 border-green-400' : ''
                        }`}
                      >
                        <div className="col-span-3 text-white">
                          <div className="flex items-center space-x-2">
                            <span className="text-gray-500">#{String(index + 1).padStart(2, '0')}</span>
                            <span>{playlist.name}</span>
                            {isActive && (
                              <span className="text-xs bg-green-400 text-black px-2 py-1 font-bold">
                                LIVE
                              </span>
                            )}
                          </div>
                        </div>
                        <div className="col-span-3 text-gray-300 truncate">
                          {playlist.description || '[NO_DESCRIPTION]'}
                        </div>
                        <div className="col-span-1 text-cyan-400">
                          {playlist.song_count.toString().padStart(3, '0')}
                        </div>
                        <div className="col-span-2 text-gray-400 text-xs">
                          {formatDate(playlist.created_at)}
                        </div>
                        <div className="col-span-1">
                          {isActive ? (
                            <span className="text-green-400 font-bold">ACTIVE</span>
                          ) : (
                            <span className="text-gray-500">STANDBY</span>
                          )}
                        </div>
                        <div className="col-span-2">
                          {!isActive && (
                            <button
                              onClick={() => handleSetActivePlaylist(playlist.id)}
                              disabled={setActivePlaylistMutation.isPending}
                              className="bg-black border border-cyan-400 hover:bg-cyan-400 hover:text-black text-cyan-400 font-mono py-1 px-3 text-xs transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                              {setActivePlaylistMutation.isPending ? (
                                <span className="flex items-center space-x-1">
                                  <div className="w-3 h-3 border border-cyan-400 border-t-transparent animate-spin"></div>
                                  <span>EXEC...</span>
                                </span>
                              ) : (
                                'ACTIVATE'
                              )}
                            </button>
                          )}
                          {isActive && (
                            <span className="text-green-400 font-mono text-xs">
                              [BROADCASTING]
                            </span>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
                
                {/* Footer */}
                <div className="bg-gray-800 border-t border-gray-600 p-3">
                  <div className="text-gray-400 font-mono text-xs">
                    <span className="text-green-400">TOTAL_PLAYLISTS:</span> {playlists.length} │ 
                    <span className="text-green-400"> ACTIVE:</span> {playlists.filter(p => queueInfo?.Playlist?.id?.toString() === p.id).length} │ 
                    <span className="text-green-400"> LAST_SCAN:</span> {new Date().toLocaleTimeString()}
                  </div>
                </div>
              </div>
            )}

            {playlists && playlists.length === 0 && (
              <div className="text-center py-12 bg-gray-900/30 border border-gray-700">
                <div className="text-yellow-400 font-mono text-lg mb-2">[WARNING]</div>
                <p className="text-gray-400 font-mono">NO PLAYLIST RECORDS FOUND IN DATABASE</p>
                <p className="text-gray-500 font-mono text-sm mt-1">SYSTEM REQUIRES AT LEAST ONE PLAYLIST TO OPERATE</p>
              </div>
            )}
          </div>
        </div>

        {/* System Status */}
        <div className="bg-black border border-gray-700">
          <div className="border-b border-gray-700 p-4">
            <h2 className="text-green-400 font-mono font-bold text-lg tracking-wider">
              ► SYSTEM_STATUS_MONITOR
            </h2>
          </div>
          <div className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {/* Authentication Status */}
              <div className="bg-gray-900 border border-gray-600 p-4">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-green-400 font-mono text-sm font-bold">AUTH_MODULE</h3>
                  <div className="w-2 h-2 bg-green-400"></div>
                </div>
                <div className="space-y-2 font-mono text-xs">
                  <div>
                    <span className="text-gray-400">USER_ID:</span>
                    <span className="text-white ml-2">{user?.username}</span>
                  </div>
                  <div>
                    <span className="text-gray-400">STATUS:</span>
                    <span className="text-green-400 ml-2">AUTHENTICATED</span>
                  </div>
                  <div>
                    <span className="text-gray-400">LEVEL:</span>
                    <span className="text-yellow-400 ml-2">ADMIN</span>
                  </div>
                </div>
              </div>

              {/* Connection Status */}
              <div className="bg-gray-900 border border-gray-600 p-4">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-green-400 font-mono text-sm font-bold">CONNECTION</h3>
                  <div className="w-2 h-2 bg-green-400"></div>
                </div>
                <div className="space-y-2 font-mono text-xs">
                  <div>
                    <span className="text-gray-400">WEBSOCKET:</span>
                    <span className="text-green-400 ml-2">CONNECTED</span>
                  </div>
                  <div>
                    <span className="text-gray-400">DATABASE:</span>
                    <span className="text-green-400 ml-2">ONLINE</span>
                  </div>
                  <div>
                    <span className="text-gray-400">API:</span>
                    <span className="text-green-400 ml-2">OPERATIONAL</span>
                  </div>
                </div>
              </div>

              {/* Current Playlist */}
              <div className="bg-gray-900 border border-gray-600 p-4">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-green-400 font-mono text-sm font-bold">ACTIVE_QUEUE</h3>
                  <div className={`w-2 h-2 ${queueInfo?.Playlist ? 'bg-green-400' : 'bg-red-400'}`}></div>
                </div>
                <div className="space-y-2 font-mono text-xs">
                  <div>
                    <span className="text-gray-400">PLAYLIST:</span>
                    <span className="text-white ml-2">
                      {queueInfo?.Playlist?.name || '[NONE]'}
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-400">TRACK:</span>
                    <span className="text-cyan-400 ml-2">
                      {getCurrentSong(queueInfo)?.title || '[NO_TRACK]'}
                    </span>
                  </div>
                  <div>
                    <span className="text-gray-400">QUEUE:</span>
                    <span className="text-yellow-400 ml-2">
                      {queueInfo?.Queue?.length || 0} TRACKS
                    </span>
                  </div>
                </div>
              </div>
            </div>

            {/* Terminal Footer */}
            <div className="mt-6 pt-4 border-t border-gray-700">
              <div className="bg-black border border-gray-600 p-3">
                <div className="font-mono text-xs text-green-400">
                  <div className="mb-1">
                    <span className="text-gray-400">admin@go-radio:~$</span> systemctl status go-radio
                  </div>
                  <div className="text-green-400 ml-4">
                    ● go-radio.service - GO Radio Broadcasting System
                  </div>
                  <div className="text-green-400 ml-4">
                    Loaded: loaded (/etc/systemd/system/go-radio.service; enabled; preset: enabled)
                  </div>
                  <div className="text-green-400 ml-4">
                    Active: active (running) since {new Date().toLocaleDateString()} {new Date().toLocaleTimeString()}
                  </div>
                  <div className="text-green-400 ml-4">
                    Process: {Math.floor(Math.random() * 9000) + 1000} (ExecStart=/usr/bin/go-radio-server)
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export const AdminPage: React.FC = () => {
  return (
    <ProtectedRoute>
      <AdminPageContent />
    </ProtectedRoute>
  );
}; 