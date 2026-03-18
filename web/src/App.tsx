import { useState, useCallback, useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AppLayout } from './components/AppLayout';
import { PhaseDashboard } from './pages/PhaseDashboard';
import { VideoList } from './pages/VideoList';
import { VideoDetail } from './pages/VideoDetail';
import { SearchResults } from './pages/SearchResults';
import { AnalyzeTitles } from './pages/AnalyzeTitles';
import { AnalyzeTiming } from './pages/AnalyzeTiming';
import { AskMeAnything } from './pages/AskMeAnything';
import { AuthScreen } from './pages/AuthScreen';
import { ApiError } from './api/client';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        if (error instanceof ApiError && error.status === 401) return false;
        return failureCount < 2;
      },
      staleTime: 30_000,
    },
  },
});

export default function App() {
  const [needsAuth, setNeedsAuth] = useState(
    () => !localStorage.getItem('api_token'),
  );

  useEffect(() => {
    queryClient.setDefaultOptions({
      queries: {
        ...queryClient.getDefaultOptions().queries,
        retry: (failureCount, error) => {
          if (error instanceof ApiError && error.status === 401) {
            setNeedsAuth(true);
            return false;
          }
          return failureCount < 2;
        },
      },
    });
  }, []);

  const handleAuthenticated = useCallback(() => {
    setNeedsAuth(false);
    queryClient.invalidateQueries();
  }, []);

  if (needsAuth) {
    return <AuthScreen onAuthenticated={handleAuthenticated} />;
  }

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route element={<AppLayout />}>
            <Route path="/" element={<PhaseDashboard />} />
            <Route path="/phases/:phaseId" element={<VideoList />} />
            <Route path="/search" element={<SearchResults />} />
            <Route path="/analyze/titles" element={<AnalyzeTitles />} />
            <Route path="/analyze/timing" element={<AnalyzeTiming />} />
            <Route path="/ama" element={<AskMeAnything />} />
            <Route
              path="/videos/:category/:videoName"
              element={<VideoDetail />}
            />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
