import { useReactions } from "../contexts/ReactionContext";

export const AnimatedEmotes = () => {
  const { reactions } = useReactions();

  return (
    <div className="fixed inset-0 pointer-events-none z-50 overflow-hidden">
      {reactions.map((reaction) => (
        <div
          key={reaction.id}
          className="absolute animate-emote-float"
          style={{
            left: `${reaction.x}%`,
            top: `${reaction.y}%`,
            animationDelay: "0s",
            animationDuration: "3s",
          }}
        >
          <div className="text-4xl drop-shadow-lg">{reaction.emoji}</div>
        </div>
      ))}
    </div>
  );
};
