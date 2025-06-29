import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
} from "react";
import type { ReactNode } from "react";
import { toast } from "react-hot-toast";
import api from "../lib/axios";

interface User {
  username: string;
}

interface AuthContextType {
  // State
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  isLoginLoading: boolean;

  // Actions
  login: (username: string, password: string) => Promise<boolean>;
  logout: () => void;
  refreshToken: () => Promise<boolean>;
  checkAuth: () => Promise<void>;
  
  // Computed
  isAuthRequired: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

const TOKEN_KEY = "auth_token";
const USER_KEY = "auth_user";

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoginLoading, setIsLoginLoading] = useState(false);

  // Check if token exists in localStorage on mount
  useEffect(() => {
    const storedToken = localStorage.getItem(TOKEN_KEY);
    const storedUser = localStorage.getItem(USER_KEY);

    if (storedToken && storedUser) {
      try {
        setToken(storedToken);
        setUser(JSON.parse(storedUser));
        // Verify token is still valid
        checkAuth();
      } catch (error) {
        console.error("Error parsing stored user data:", error);
        logout();
      }
    } else {
      setIsLoading(false);
    }
  }, []);

  // Set up axios interceptor to include auth header
  useEffect(() => {
    const requestInterceptor = api.interceptors.request.use(
      (config) => {
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    const responseInterceptor = api.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401 && token) {
          // Token is invalid, logout user
          logout();
          toast.error("Session expired. Please login again.");
        }
        return Promise.reject(error);
      }
    );

    return () => {
      api.interceptors.request.eject(requestInterceptor);
      api.interceptors.response.eject(responseInterceptor);
    };
  }, [token]);

  const login = useCallback(async (username: string, password: string): Promise<boolean> => {
    setIsLoginLoading(true);
    try {
      const response = await api.post("/auth/login", {
        username,
        password,
      });

      const { token: newToken, message } = response.data;
      
      if (newToken) {
        const newUser = { username };
        
        // Store in localStorage
        localStorage.setItem(TOKEN_KEY, newToken);
        localStorage.setItem(USER_KEY, JSON.stringify(newUser));
        
        // Update state
        setToken(newToken);
        setUser(newUser);
        
        toast.success(message || "Login successful!");
        return true;
      }
      
      return false;
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || "Login failed";
      toast.error(errorMessage);
      return false;
    } finally {
      setIsLoginLoading(false);
    }
  }, []);

  const logout = useCallback(() => {
    // Clear localStorage
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    
    // Clear state
    setToken(null);
    setUser(null);
    
    toast.success("Logged out successfully");
  }, []);

  const refreshToken = useCallback(async (): Promise<boolean> => {
    if (!token) return false;

    try {
      const response = await api.post("/auth/refresh", {
        token,
      });

      const { token: newToken } = response.data;
      
      if (newToken) {
        localStorage.setItem(TOKEN_KEY, newToken);
        setToken(newToken);
        return true;
      }
      
      return false;
    } catch (error) {
      console.error("Token refresh failed:", error);
      logout();
      return false;
    }
  }, [token, logout]);

  const checkAuth = useCallback(async () => {
    if (!token) {
      setIsLoading(false);
      return;
    }

    try {
      const response = await api.get("/auth/me");
      const { username } = response.data;
      
      // Update user info if successful
      setUser({ username });
    } catch (error) {
      console.error("Auth check failed:", error);
      logout();
    } finally {
      setIsLoading(false);
    }
  }, [token, logout]);

  // Auto-refresh token before expiration (every 23 hours if token expires in 24 hours)
  useEffect(() => {
    if (!token) return;

    const refreshInterval = setInterval(() => {
      refreshToken();
    }, 23 * 60 * 60 * 1000); // 23 hours

    return () => clearInterval(refreshInterval);
  }, [token, refreshToken]);

  const isAuthenticated = !!user && !!token;
  const isAuthRequired = !isAuthenticated; // Can be used to determine if login is required

  const value: AuthContextType = {
    user,
    token,
    isAuthenticated,
    isLoading,
    isLoginLoading,
    login,
    logout,
    refreshToken,
    checkAuth,
    isAuthRequired,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}; 