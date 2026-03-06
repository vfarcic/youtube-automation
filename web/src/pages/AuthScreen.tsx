import { useState } from 'react';

interface AuthScreenProps {
  onAuthenticated: () => void;
}

export function AuthScreen({ onAuthenticated }: AuthScreenProps) {
  const [token, setToken] = useState('');

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (token.trim()) {
      localStorage.setItem('api_token', token.trim());
      onAuthenticated();
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-900">
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-sm rounded-lg bg-gray-800 p-6 shadow-lg shadow-black/30"
      >
        <h1 className="text-lg font-bold text-gray-100 mb-4">
          API Authentication
        </h1>
        <p className="text-sm text-gray-400 mb-4">
          Enter your API bearer token to continue.
        </p>
        <input
          type="password"
          value={token}
          onChange={(e) => setToken(e.target.value)}
          placeholder="Bearer token"
          className="w-full rounded border border-gray-600 bg-gray-700 text-gray-100 px-3 py-2 text-sm mb-3 focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-gray-500"
        />
        <button
          type="submit"
          className="w-full rounded bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          Connect
        </button>
      </form>
    </div>
  );
}
