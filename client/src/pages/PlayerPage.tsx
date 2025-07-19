import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { RadioPlayer } from "../components/RadioPlayer";
import { useRadio } from "../contexts/RadioContext";

export const PlayerPage = () => {
  const navigate = useNavigate();
  const { isUserInteracted, isAudioContextReady, startPlaybackOnMount } = useRadio();

  // Redirect to intro if user hasn't interacted yet
  useEffect(() => {
    if (!isUserInteracted || !isAudioContextReady) {
      navigate("/");
      return;
    }

    // Start playback when player mounts
    startPlaybackOnMount();
  }, [isUserInteracted, isAudioContextReady, navigate, startPlaybackOnMount]);

  // Don't render anything if not ready (will redirect)
  if (!isUserInteracted || !isAudioContextReady) {
    return null;
  }

  return (
    <div className="w-full flex items-center justify-center p-2 sm:p-4">
      <RadioPlayer />
    </div>
  );
};