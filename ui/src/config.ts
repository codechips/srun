declare global {
  interface Window {
    APP_BASE_PATH?: string;
  }
}

const rawAppBasePath = typeof window !== 'undefined' ? window.APP_BASE_PATH || "/" : "/";

export const API_BASE_PREFIX = rawAppBasePath === "/" ? "" : rawAppBasePath;

export const getApiUrl = (path: string): string => {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  return `${API_BASE_PREFIX}${normalizedPath}`;
};

export const getWsUrl = (path: string): string => {
  const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const host = window.location.host;
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  return `${wsProtocol}//${host}${API_BASE_PREFIX}${normalizedPath}`;
};
