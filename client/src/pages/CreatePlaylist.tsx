import { useState } from "react";
import { useForm } from "@tanstack/react-form";
import { z } from "zod";
import { useMutation, useQuery } from "@tanstack/react-query";
import {
  MagnifyingGlassIcon,
  PlusIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import { toast } from "react-hot-toast";
import { useNavigate } from "react-router-dom";
import { isAxiosError } from "axios";
import api from "../lib/axios";
import { ProtectedRoute } from "../components/ProtectedRoute";

// Define the form schema
const playlistSchema = z.object({
  name: z
    .string()
    .min(1, "Playlist name is required")
    .max(100, "Name is too long"),
  description: z.string().max(500, "Description is too long").default(""),
});

type FormValues = z.infer<typeof playlistSchema>;

interface Song {
  id: string;
  title: string;
  description: string;
  thumbnail: string;
  duration: string;
}

// Custom hook for YouTube search
function useYouTubeSearch(query: string) {
  return useQuery({
    queryKey: ["youtube-search", query],
    queryFn: async () => {
      if (!query.trim()) return [];
      const response = await api.get<Song[]>(
        `/youtube/search?q=${encodeURIComponent(query)}`
      );
      return response.data;
    },
  });
}

export function CreatePlaylist() {
  const navigate = useNavigate();
  const [searchInput, setSearchInput] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedSongs, setSelectedSongs] = useState<Song[]>([]);

  // Initialize form with TanStack Form
  const form = useForm({
    defaultValues: {
      name: "",
      description: "",
    } as FormValues,
    onSubmit: async ({ value }) => {
      if (selectedSongs.length === 0) {
        toast.error("Please add at least one song to the playlist");
        return;
      }

      await createPlaylist.mutateAsync({
        ...value,
        songs: selectedSongs.map((song) => song.id),
      });
    },
  });

  // Use React Query for YouTube search
  const {
    data: searchResults = [],
    isLoading: isSearching,
    error: searchError,
  } = useYouTubeSearch(searchQuery);

  const createPlaylist = useMutation({
    mutationFn: async (playlist: FormValues & { songs: string[] }) => {
      const response = await api.post("/playlists", playlist);
      return response.data;
    },
    onSuccess: () => {
      toast.success("Playlist created successfully");
      navigate("/playlists");
    },
    onError: (error: unknown) => {
      if (isAxiosError(error)) {
        toast.error(
          error.response?.data?.message || "Failed to create playlist"
        );
      } else {
        toast.error("Failed to create playlist");
      }
    },
  });

  // Handle search form submission
  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setSearchQuery(searchInput);
  };

  // Handle search input change
  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchInput(e.target.value);
  };

  const addSong = (song: Song) => {
    if (!selectedSongs.find((s) => s.id === song.id)) {
      setSelectedSongs([...selectedSongs, song]);
    }
  };

  const removeSong = (songId: string) => {
    setSelectedSongs(selectedSongs.filter((song) => song.id !== songId));
  };

  // Helper function to format ISO 8601 duration
  const formatDuration = (duration: string) => {
    const match = duration.match(/PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?/);
    if (!match) return duration;

    const [, hours, minutes, seconds] = match;
    const parts = [];

    if (hours) parts.push(`${hours}h`);
    if (minutes) parts.push(`${minutes}m`);
    if (seconds) parts.push(`${seconds}s`);

    return parts.join(" ") || duration;
  };

  return (
    <ProtectedRoute>
    <div className="max-w-4xl mx-auto p-6">
      <h1 className="text-2xl font-mono font-bold text-white mb-8 tracking-wider">
        [CREATE_PLAYLIST]
      </h1>

      {/* Search form - moved outside the playlist form */}
      <div className="mb-8">
        <h2 className="text-lg font-mono font-bold text-white mb-4 tracking-wider">
          [ADD_SONGS]
        </h2>
        <form onSubmit={handleSearchSubmit} className="mb-4">
          <div className="flex gap-2">
            <input
              type="text"
              value={searchInput}
              onChange={handleSearchChange}
              placeholder="Search YouTube..."
              className="flex-1 px-4 py-2.5 bg-black border border-gray-700 text-white placeholder:text-gray-500 text-sm font-mono focus:border-white focus:outline-none"
            />
            <button
              type="submit"
              disabled={isSearching || !searchInput.trim()}
              className="inline-flex items-center px-4 py-2 bg-black border border-white text-white font-mono text-sm hover:bg-white hover:text-black transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <MagnifyingGlassIcon className="h-5 w-5 mr-2" />
              {isSearching ? "[SEARCHING...]" : "[SEARCH]"}
            </button>
          </div>
        </form>

        {searchError && (
          <div className="mb-4 p-4 bg-black border border-gray-700">
            <p className="text-sm text-gray-500 font-mono">
              {searchError instanceof Error
                ? searchError.message
                : "Failed to search YouTube"}
            </p>
          </div>
        )}

        {searchResults.length > 0 && (
          <div className="space-y-2 mb-6">
            <h3 className="text-sm font-mono font-bold text-white tracking-wider">
              [SEARCH_RESULTS]
            </h3>
            {searchResults.map((song) => (
              <div
                key={song.id}
                className="flex items-center justify-between p-3 bg-black border border-gray-700 hover:border-gray-600 transition-colors"
              >
                <div className="flex items-center min-w-0 flex-1 mr-4">
                  <img
                    src={song.thumbnail}
                    alt={song.title}
                    className="w-16 h-16 object-cover border border-gray-700 flex-shrink-0"
                  />
                  <div className="ml-4 min-w-0 flex-1">
                    <h4 className="font-mono text-white truncate text-sm">
                      {song.title}
                    </h4>
                    <p className="text-xs text-gray-500 font-mono">
                      {formatDuration(song.duration)}
                    </p>
                    <p className="text-xs text-gray-600 truncate font-mono">
                      {song.description}
                    </p>
                  </div>
                </div>
                {selectedSongs.find((s) => s.id === song.id) ? (
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-white font-mono">
                      [ADDED]
                    </span>
                    <button
                      type="button"
                      onClick={() => removeSong(song.id)}
                      className="p-2 text-gray-500 hover:text-white border border-gray-700 hover:border-white transition-colors"
                      title="Remove from playlist"
                    >
                      <XMarkIcon className="h-4 w-4" />
                    </button>
                  </div>
                ) : (
                  <button
                    type="button"
                    onClick={() => addSong(song)}
                    className="p-2 text-gray-500 hover:text-white border border-gray-700 hover:border-white transition-colors"
                    title="Add to playlist"
                  >
                    <PlusIcon className="h-4 w-4" />
                  </button>
                )}
              </div>
            ))}
          </div>
        )}

        {selectedSongs.length > 0 && (
          <div className="space-y-2">
            <h3 className="text-sm font-mono font-bold text-white tracking-wider">
              [SELECTED_SONGS] ({selectedSongs.length})
            </h3>
            {selectedSongs.map((song) => (
              <div
                key={song.id}
                className="flex items-center justify-between p-3 bg-black border border-gray-700 hover:border-gray-600 transition-colors"
              >
                <div className="flex items-center min-w-0 flex-1 mr-4">
                  <img
                    src={song.thumbnail}
                    alt={song.title}
                    className="w-16 h-16 object-cover border border-gray-700 flex-shrink-0"
                  />
                  <div className="ml-4 min-w-0 flex-1">
                    <h4 className="font-mono text-white truncate text-sm">
                      {song.title}
                    </h4>
                    <p className="text-xs text-gray-500 font-mono">
                      {formatDuration(song.duration)}
                    </p>
                  </div>
                </div>
                <button
                  type="button"
                  onClick={() => removeSong(song.id)}
                  className="p-2 text-gray-500 hover:text-white border border-gray-700 hover:border-white transition-colors"
                  title="Remove from playlist"
                >
                  <XMarkIcon className="h-4 w-4" />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Playlist creation form */}
      <form
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          void form.handleSubmit();
        }}
        className="space-y-6"
      >
        <div>
          <label
            htmlFor="name"
            className="block text-sm font-mono font-bold text-white tracking-wider"
          >
            [PLAYLIST_NAME]
          </label>
          <form.Field
            name="name"
            validators={{
              onChange: ({ value }) => {
                const result = playlistSchema.shape.name.safeParse(value);
                return result.success
                  ? undefined
                  : result.error.errors[0].message;
              },
            }}
            children={(field) => (
              <>
                <input
                  id="name"
                  type="text"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  placeholder="Enter playlist name"
                  className={`mt-1 block w-full px-4 py-2.5 bg-black border text-white font-mono text-sm placeholder:text-gray-500 focus:outline-none ${
                    field.state.meta.errors.length > 0
                      ? "border-gray-500"
                      : "border-gray-700 focus:border-white"
                  }`}
                />
                {field.state.meta.errors.length > 0 && (
                  <p className="mt-1 text-xs text-gray-500 font-mono">
                    {field.state.meta.errors[0]}
                  </p>
                )}
              </>
            )}
          />
        </div>

        <div>
          <label
            htmlFor="description"
            className="block text-sm font-mono font-bold text-white tracking-wider"
          >
            [DESCRIPTION]
          </label>
          <form.Field
            name="description"
            validators={{
              onChange: ({ value }) => {
                const result =
                  playlistSchema.shape.description.safeParse(value);
                return result.success
                  ? undefined
                  : result.error.errors[0].message;
              },
            }}
            children={(field) => (
              <>
                <textarea
                  id="description"
                  rows={3}
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  placeholder="Enter playlist description (optional)"
                  className={`mt-1 block w-full px-4 py-2.5 bg-black border text-white font-mono text-sm placeholder:text-gray-500 resize-none focus:outline-none ${
                    field.state.meta.errors.length > 0
                      ? "border-gray-500"
                      : "border-gray-700 focus:border-white"
                  }`}
                />
                {field.state.meta.errors.length > 0 && (
                  <p className="mt-1 text-xs text-gray-500 font-mono">
                    {field.state.meta.errors[0]}
                  </p>
                )}
              </>
            )}
          />
        </div>

        <div className="flex justify-end">
          <button
            type="submit"
            disabled={
              form.state.isSubmitting ||
              createPlaylist.isPending ||
              selectedSongs.length === 0
            }
            className="inline-flex items-center px-4 py-2 bg-black border border-white text-white font-mono text-sm hover:bg-white hover:text-black transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {form.state.isSubmitting || createPlaylist.isPending
              ? "[CREATING...]"
              : "[CREATE_PLAYLIST]"}
          </button>
        </div>
      </form>
    </div>
    </ProtectedRoute>
  );
}
