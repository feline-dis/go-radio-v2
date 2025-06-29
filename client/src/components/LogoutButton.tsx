import React from "react";
import { useAuth } from "../contexts/AuthContext";
import { ArrowRightOnRectangleIcon } from "@heroicons/react/24/outline";

interface LogoutButtonProps {
  className?: string;
  variant?: "icon" | "text" | "both";
}

export const LogoutButton: React.FC<LogoutButtonProps> = ({
  className = "",
  variant = "both",
}) => {
  const { logout, user } = useAuth();

  const handleLogout = () => {
    logout();
  };

  return (
    <button
      onClick={handleLogout}
      className={`flex items-center space-x-2 px-3 py-2 rounded-lg text-gray-300 hover:text-white hover:bg-gray-800 transition-colors ${className}`}
      title={`Logout ${user?.username || ""}`}
    >
      {(variant === "icon" || variant === "both") && (
        <ArrowRightOnRectangleIcon className="h-5 w-5" />
      )}
      {(variant === "text" || variant === "both") && (
        <span className="font-mono text-sm">
          {variant === "text" ? `[LOGOUT_${user?.username?.toUpperCase()}]` : "[LOGOUT]"}
        </span>
      )}
    </button>
  );
}; 